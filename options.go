package squad

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring a Client.
type Option func(*config)

// config holds defaults before the Client is constructed.
type config struct {
	baseURL         string
	httpClient      *http.Client
	timeout         time.Duration
	userAgent       string
	logger          Logger
	beforeRequest   func(*http.Request)
	afterResponse   func(*http.Request, *http.Response, time.Duration)
	autoIdempotency bool
}

// WithSandbox configures the client to use Squad's sandbox environment.
// This is also inferred automatically when the secret key starts with "sandbox_sk_".
func WithSandbox() Option {
	return func(c *config) {
		c.baseURL = SandboxBaseURL
	}
}

// WithProduction explicitly sets the production base URL.
func WithProduction() Option {
	return func(c *config) {
		c.baseURL = ProductionBaseURL
	}
}

// WithBaseURL sets a custom base URL. Useful for proxies, local mocks, or testing.
// Must include scheme and host with no trailing slash: "https://my-proxy.example.com"
func WithBaseURL(url string) Option {
	return func(c *config) {
		c.baseURL = url
	}
}

// WithHTTPClient replaces the default *http.Client.
// Use this to inject transports with retry logic, tracing, or custom TLS config.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) {
		c.httpClient = hc
	}
}

// WithTimeout sets the HTTP request timeout, overriding the default 30-second timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

// WithUserAgent appends a custom string to the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *config) {
		c.userAgent = ua
	}
}

// WithLogger attaches a Logger to the client for request/response logging.
// By default all logging is discarded. Use StdLogger() for development output.
//
//	client := squad.New(key, squad.WithLogger(squad.StdLogger()))
func WithLogger(l Logger) Option {
	return func(c *config) {
		c.logger = l
	}
}

// WithBeforeRequest registers a hook called before every HTTP request is sent.
// Use this for custom header injection, request signing, or distributed tracing.
//
//	squad.WithBeforeRequest(func(req *http.Request) {
//	    req.Header.Set("X-Correlation-ID", getCorrelationID())
//	})
func WithBeforeRequest(fn func(*http.Request)) Option {
	return func(c *config) {
		c.beforeRequest = fn
	}
}

// WithAfterResponse registers a hook called after every HTTP response is received.
// duration is the total round-trip time. Use this for metrics, tracing, or audit logging.
//
//	squad.WithAfterResponse(func(req *http.Request, resp *http.Response, d time.Duration) {
//	    metrics.RecordLatency("squad", req.URL.Path, resp.StatusCode, d)
//	})
func WithAfterResponse(fn func(*http.Request, *http.Response, time.Duration)) Option {
	return func(c *config) {
		c.afterResponse = fn
	}
}

// WithAutoIdempotency automatically generates and attaches an X-Idempotency-Key
// header to every POST request that does not already have one set via WithIdempotencyKey.
// This protects against accidental duplicate charges from network-level retries.
func WithAutoIdempotency() Option {
	return func(c *config) {
		c.autoIdempotency = true
	}
}
