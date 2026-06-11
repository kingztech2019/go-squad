package squad

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// apiResponse is the envelope Squad wraps all responses in.
// Format: { "status": 200, "message": "success", "data": { ... } }.
type apiResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// do executes a JSON HTTP request and decodes the Squad envelope response.
// method is the HTTP verb (GET, POST, PUT, PATCH, DELETE).
// path is the API path without the base URL (e.g. "/transaction/initiate").
// body is the request payload, JSON-encoded if non-nil. Pass nil for GET/DELETE.
// out is a pointer to the Go type the Data field should decode into. Pass nil to ignore Data.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("squad: marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := c.buildRequest(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Inject idempotency key on POST requests.
	if method == http.MethodPost {
		if key := idempotencyKeyFromCtx(ctx); key != "" {
			req.Header.Set("X-Idempotency-Key", key)
		} else if c.autoIdempotency {
			if key, genErr := GenerateIdempotencyKey(); genErr == nil {
				req.Header.Set("X-Idempotency-Key", key)
			}
		}
	}

	if c.beforeRequest != nil {
		c.beforeRequest(req)
	}

	c.logger.Info("squad request", "method", method, "path", path)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("squad request failed",
			"method", method, "path", path,
			"duration_ms", duration.Milliseconds(), "error", err)
		return fmt.Errorf("squad: execute request: %w", err)
	}

	if c.afterResponse != nil {
		c.afterResponse(req, resp, duration)
	}

	parseErr := c.parseResponse(resp, out)
	if parseErr != nil {
		c.logger.Error("squad response error",
			"method", method, "path", path,
			"http_status", resp.StatusCode,
			"duration_ms", duration.Milliseconds(), "error", parseErr)
	} else {
		c.logger.Info("squad response",
			"method", method, "path", path,
			"http_status", resp.StatusCode,
			"duration_ms", duration.Milliseconds())
	}
	return parseErr
}

// doGet executes a GET request with optional URL query parameters.
func (c *Client) doGet(ctx context.Context, path string, params url.Values, out any) error {
	fullPath := path
	if len(params) > 0 {
		fullPath = path + "?" + params.Encode()
	}
	return c.do(ctx, http.MethodGet, fullPath, nil, out)
}

// doMultipart executes a multipart/form-data POST request.
// fields are plain string form fields. fileField is the form field name for the file upload.
// fileData is the raw file bytes; fileName is the filename for Content-Disposition.
func (c *Client) doMultipart(ctx context.Context, path string, fields map[string]string, fileField, fileName string, fileData []byte, out any) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return fmt.Errorf("squad: write multipart field %q: %w", k, err)
		}
	}

	fw, err := mw.CreateFormFile(fileField, fileName)
	if err != nil {
		return fmt.Errorf("squad: create form file: %w", err)
	}
	if _, err := fw.Write(fileData); err != nil {
		return fmt.Errorf("squad: write file data: %w", err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("squad: close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("squad: build multipart request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	if key := idempotencyKeyFromCtx(ctx); key != "" {
		req.Header.Set("X-Idempotency-Key", key)
	} else if c.autoIdempotency {
		if key, genErr := GenerateIdempotencyKey(); genErr == nil {
			req.Header.Set("X-Idempotency-Key", key)
		}
	}

	if c.beforeRequest != nil {
		c.beforeRequest(req)
	}

	c.logger.Info("squad request", "method", "POST", "path", path)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("squad request failed",
			"method", "POST", "path", path,
			"duration_ms", duration.Milliseconds(), "error", err)
		return fmt.Errorf("squad: execute multipart request: %w", err)
	}

	if c.afterResponse != nil {
		c.afterResponse(req, resp, duration)
	}

	parseErr := c.parseResponse(resp, out)
	if parseErr != nil {
		c.logger.Error("squad response error",
			"method", "POST", "path", path,
			"http_status", resp.StatusCode,
			"duration_ms", duration.Milliseconds(), "error", parseErr)
	} else {
		c.logger.Info("squad response",
			"method", "POST", "path", path,
			"http_status", resp.StatusCode,
			"duration_ms", duration.Milliseconds())
	}
	return parseErr
}

// buildRequest constructs an *http.Request with all required Squad headers.
func (c *Client) buildRequest(ctx context.Context, method, fullURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("squad: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	return req, nil
}

// parseResponse decodes a Squad envelope response.
// Returns *Error on non-2xx envelope status or HTTP error status.
func (c *Client) parseResponse(resp *http.Response, out any) error {
	defer resp.Body.Close() //nolint:errcheck

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("squad: read response body: %w", err)
	}

	var envelope apiResponse
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return &Error{
			HTTPStatus: resp.StatusCode,
			Status:     resp.StatusCode,
			Message:    fmt.Sprintf("squad: non-JSON response (http %d): %s", resp.StatusCode, string(rawBody)),
		}
	}

	if envelope.Status < 200 || envelope.Status > 299 {
		return &Error{
			HTTPStatus: resp.StatusCode,
			Status:     envelope.Status,
			Message:    envelope.Message,
		}
	}

	if out != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("squad: decode response data: %w", err)
		}
	}
	return nil
}
