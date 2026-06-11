package squad

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// WebhookRouter is an http.Handler that validates Squad webhook signatures and
// dispatches each event to its registered typed handler function.
//
// Register handlers with the fluent On* methods, then mount the router as an HTTP handler:
//
//	router := squad.NewWebhookRouter(os.Getenv("SQUAD_SECRET_KEY")).
//	    OnTransactionSuccess(handlePayment).
//	    OnVirtualAccountCredit(handleCredit).
//	    OnDisputeOpened(handleDispute)
//
//	http.Handle("/webhook/squad", router)
type WebhookRouter struct {
	secret string

	onTransactionSuccess   func(context.Context, *WebhookTransactionBody) error
	onTransactionFailed    func(context.Context, *WebhookTransactionBody) error
	onVirtualAccountCredit func(context.Context, *WebhookVirtualAccountBody) error
	onTransferSuccess      func(context.Context, *WebhookTransferBody) error
	onTransferFailed       func(context.Context, *WebhookTransferBody) error
	onTransferReversed     func(context.Context, *WebhookTransferBody) error
	onDisputeOpened        func(context.Context, *WebhookDisputeBody) error
	onDisputeResolved      func(context.Context, *WebhookDisputeBody) error
	onUnknown              func(context.Context, *WebhookEvent) error
	onError                func(http.ResponseWriter, *http.Request, error)
}

// NewWebhookRouter creates a new WebhookRouter that validates signatures using secret.
// secret is the same Squad secret key used to initialise the Client.
func NewWebhookRouter(secret string) *WebhookRouter {
	return &WebhookRouter{secret: secret}
}

// OnTransactionSuccess registers a handler for charge.success events.
func (r *WebhookRouter) OnTransactionSuccess(fn func(context.Context, *WebhookTransactionBody) error) *WebhookRouter {
	r.onTransactionSuccess = fn
	return r
}

// OnTransactionFailed registers a handler for charge.failed events.
func (r *WebhookRouter) OnTransactionFailed(fn func(context.Context, *WebhookTransactionBody) error) *WebhookRouter {
	r.onTransactionFailed = fn
	return r
}

// OnVirtualAccountCredit registers a handler for virtual-account.credit events.
func (r *WebhookRouter) OnVirtualAccountCredit(fn func(context.Context, *WebhookVirtualAccountBody) error) *WebhookRouter {
	r.onVirtualAccountCredit = fn
	return r
}

// OnTransferSuccess registers a handler for transfer.success events.
func (r *WebhookRouter) OnTransferSuccess(fn func(context.Context, *WebhookTransferBody) error) *WebhookRouter {
	r.onTransferSuccess = fn
	return r
}

// OnTransferFailed registers a handler for transfer.failed events.
func (r *WebhookRouter) OnTransferFailed(fn func(context.Context, *WebhookTransferBody) error) *WebhookRouter {
	r.onTransferFailed = fn
	return r
}

// OnTransferReversed registers a handler for transfer.reversed events.
func (r *WebhookRouter) OnTransferReversed(fn func(context.Context, *WebhookTransferBody) error) *WebhookRouter {
	r.onTransferReversed = fn
	return r
}

// OnDisputeOpened registers a handler for dispute.opened events.
func (r *WebhookRouter) OnDisputeOpened(fn func(context.Context, *WebhookDisputeBody) error) *WebhookRouter {
	r.onDisputeOpened = fn
	return r
}

// OnDisputeResolved registers a handler for dispute.resolved events.
func (r *WebhookRouter) OnDisputeResolved(fn func(context.Context, *WebhookDisputeBody) error) *WebhookRouter {
	r.onDisputeResolved = fn
	return r
}

// OnUnknown registers a fallback handler for unrecognised or future event types.
// The raw *WebhookEvent is passed so the caller can inspect event.Body directly.
func (r *WebhookRouter) OnUnknown(fn func(context.Context, *WebhookEvent) error) *WebhookRouter {
	r.onUnknown = fn
	return r
}

// OnError registers a custom error handler called when signature validation fails,
// JSON parsing fails, or a dispatched handler returns an error.
// The default behaviour is to write an HTTP 400 or 403 response.
func (r *WebhookRouter) OnError(fn func(http.ResponseWriter, *http.Request, error)) *WebhookRouter {
	r.onError = fn
	return r
}

// ServeHTTP implements http.Handler. It reads the request body, validates the
// x-squad-signature header, parses the event, and dispatches to the registered handler.
// Returns 403 on invalid signature, 400 on parse errors, 500 if a handler returns an error,
// and 200 on success.
func (r *WebhookRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.handleErr(w, req, fmt.Errorf("squad webhook: read body: %w", err))
		return
	}

	sig := req.Header.Get("x-squad-signature")
	event, err := ParseWebhook(body, sig, r.secret)
	if err != nil {
		if errors.Is(err, ErrInvalidSignature) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		r.handleErr(w, req, err)
		return
	}

	if handlerErr := r.dispatch(req.Context(), event); handlerErr != nil {
		r.handleErr(w, req, handlerErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// dispatch routes a parsed event to the appropriate registered handler.
func (r *WebhookRouter) dispatch(ctx context.Context, event *WebhookEvent) error {
	parsed, err := event.ParseBody()
	if err != nil {
		return err
	}

	switch event.Event {
	case EventTransactionSuccess:
		if r.onTransactionSuccess != nil {
			return r.onTransactionSuccess(ctx, parsed.(*WebhookTransactionBody))
		}
	case EventTransactionFailed:
		if r.onTransactionFailed != nil {
			return r.onTransactionFailed(ctx, parsed.(*WebhookTransactionBody))
		}
	case EventVirtualAccountCredit:
		if r.onVirtualAccountCredit != nil {
			return r.onVirtualAccountCredit(ctx, parsed.(*WebhookVirtualAccountBody))
		}
	case EventTransferSuccess:
		if r.onTransferSuccess != nil {
			return r.onTransferSuccess(ctx, parsed.(*WebhookTransferBody))
		}
	case EventTransferFailed:
		if r.onTransferFailed != nil {
			return r.onTransferFailed(ctx, parsed.(*WebhookTransferBody))
		}
	case EventTransferReversed:
		if r.onTransferReversed != nil {
			return r.onTransferReversed(ctx, parsed.(*WebhookTransferBody))
		}
	case EventDisputeOpened:
		if r.onDisputeOpened != nil {
			return r.onDisputeOpened(ctx, parsed.(*WebhookDisputeBody))
		}
	case EventDisputeResolved:
		if r.onDisputeResolved != nil {
			return r.onDisputeResolved(ctx, parsed.(*WebhookDisputeBody))
		}
	default:
		if r.onUnknown != nil {
			return r.onUnknown(ctx, event)
		}
	}
	return nil
}

func (r *WebhookRouter) handleErr(w http.ResponseWriter, req *http.Request, err error) {
	if r.onError != nil {
		r.onError(w, req, err)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
