package squad

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// ParseWebhook parses and validates an inbound Squad webhook.
//
// payload is the raw HTTP request body bytes.
// signature is the value of the "x-squad-signature" header.
// secret is the merchant's Squad secret key (same key used for API calls).
//
// Returns ErrInvalidSignature if the HMAC-SHA512 signature does not match.
// Returns a parse error if the JSON payload is malformed.
func ParseWebhook(payload []byte, signature string, secret string) (*WebhookEvent, error) {
	if !VerifySignature(payload, signature, secret) {
		return nil, ErrInvalidSignature
	}
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("squad: parse webhook payload: %w", err)
	}
	return &event, nil
}

// VerifySignature checks whether the HMAC-SHA512 signature matches the payload.
// signature is the hex-encoded value from the "x-squad-signature" header.
// secret is the merchant's Squad secret key.
//
// Uses constant-time comparison to prevent timing attacks.
func VerifySignature(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
