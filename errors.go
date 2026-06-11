package squad

import (
	"errors"
	"fmt"
)

// Error represents a Squad API error response.
// It wraps both the HTTP-level status and Squad's application-level status code.
type Error struct {
	HTTPStatus int
	Status     int
	Message    string
	Code       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("squad: status %d: %s (http=%d)", e.Status, e.Message, e.HTTPStatus)
}

// Is allows errors.Is comparisons by HTTPStatus and Status fields, not pointer identity.
func (e *Error) Is(target error) bool {
	var t *Error
	if errors.As(target, &t) {
		return e.HTTPStatus == t.HTTPStatus && e.Status == t.Status
	}
	return false
}

// Sentinel errors for common API failure modes.
var (
	ErrUnauthorized = &Error{HTTPStatus: 401, Status: 401, Message: "unauthorized"}
	ErrForbidden    = &Error{HTTPStatus: 403, Status: 403, Message: "forbidden"}
	ErrBadRequest   = &Error{HTTPStatus: 400, Status: 400, Message: "bad request"}
	ErrNotFound     = &Error{HTTPStatus: 404, Status: 404, Message: "not found"}
)

// ErrInvalidSignature is returned by ParseWebhook when HMAC-SHA512 validation fails.
var ErrInvalidSignature = fmt.Errorf("squad: webhook signature validation failed")

// IsUnauthorized reports whether err is a 401 authorization failure.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsBadRequest reports whether err is a 400 validation failure.
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}

// IsNotFound reports whether err is a 404 not-found failure.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
