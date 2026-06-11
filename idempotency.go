package squad

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type ctxKey string

const idempotencyCtxKey ctxKey = "squad_idempotency_key"

// WithIdempotencyKey attaches an idempotency key to the context.
// The SDK sends it as X-Idempotency-Key on all POST requests, preventing
// duplicate charges when a request is retried after a network failure.
//
// The key should be unique per business operation and attempt.
// Store it before making the request so you can reuse it on retry.
//
//	key, _ := squad.GenerateIdempotencyKey()
//	ctx = squad.WithIdempotencyKey(ctx, "order-"+orderID+"-"+key)
//	resp, err := client.Transactions.InitiatePayment(ctx, params)
//	// On network error, retry with the SAME ctx to use the same key.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyCtxKey, key)
}

// GenerateIdempotencyKey generates a cryptographically random 32-character hex key.
// Use this when you don't have a natural business key (e.g. order ID) to use.
//
//	key, err := squad.GenerateIdempotencyKey()
//	// Store key alongside the order before initiating payment.
//	ctx = squad.WithIdempotencyKey(ctx, key)
func GenerateIdempotencyKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("squad: generate idempotency key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// idempotencyKeyFromCtx extracts the idempotency key from a context.
// Returns an empty string if no key was set.
func idempotencyKeyFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(idempotencyCtxKey).(string)
	return v
}
