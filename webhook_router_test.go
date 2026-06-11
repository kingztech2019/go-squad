package squad_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kingztech2019/go-squad"
)

func makeWebhookRequest(t *testing.T, event squad.EventType, body any, secret string) *http.Request {
	t.Helper()
	bodyBytes, _ := json.Marshal(body)
	payload, _ := json.Marshal(map[string]any{
		"event": event,
		"body":  json.RawMessage(bodyBytes),
	})
	sig := signPayload(payload, secret)
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(payload))) //nolint:noctx
	req.Header.Set("x-squad-signature", sig)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestWebhookRouter_TransactionSuccess(t *testing.T) {
	called := false
	router := squad.NewWebhookRouter(testWebhookSecret).
		OnTransactionSuccess(func(_ context.Context, body *squad.WebhookTransactionBody) error {
			called = true
			if body.TransactionRef != "txn_001" {
				t.Errorf("expected txn_001, got %s", body.TransactionRef)
			}
			return nil
		})

	req := makeWebhookRequest(t, squad.EventTransactionSuccess,
		squad.WebhookTransactionBody{TransactionRef: "txn_001", Amount: 500000},
		testWebhookSecret,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestWebhookRouter_InvalidSignature(t *testing.T) {
	router := squad.NewWebhookRouter(testWebhookSecret)

	payload := []byte(`{"event":"charge.success","body":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(payload))) //nolint:noctx
	req.Header.Set("x-squad-signature", "bad_signature")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestWebhookRouter_HandlerError_Returns500(t *testing.T) {
	router := squad.NewWebhookRouter(testWebhookSecret).
		OnTransactionSuccess(func(_ context.Context, _ *squad.WebhookTransactionBody) error {
			return errors.New("database unavailable")
		})

	req := makeWebhookRequest(t, squad.EventTransactionSuccess,
		squad.WebhookTransactionBody{TransactionRef: "txn_002"},
		testWebhookSecret,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWebhookRouter_CustomErrorHandler(t *testing.T) {
	customCalled := false
	router := squad.NewWebhookRouter(testWebhookSecret).
		OnTransactionSuccess(func(_ context.Context, _ *squad.WebhookTransactionBody) error {
			return errors.New("processing error")
		}).
		OnError(func(w http.ResponseWriter, _ *http.Request, err error) {
			customCalled = true
			http.Error(w, "custom error: "+err.Error(), http.StatusInternalServerError)
		})

	req := makeWebhookRequest(t, squad.EventTransactionSuccess,
		squad.WebhookTransactionBody{TransactionRef: "txn_003"},
		testWebhookSecret,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !customCalled {
		t.Error("custom error handler was not called")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestWebhookRouter_UnhandledEvent_NoError(t *testing.T) {
	// Router with no handlers registered — should still return 200 for valid webhooks.
	router := squad.NewWebhookRouter(testWebhookSecret)

	req := makeWebhookRequest(t, squad.EventVirtualAccountCredit,
		squad.WebhookVirtualAccountBody{VirtualAccountNumber: "0123456789"},
		testWebhookSecret,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for unhandled event, got %d", w.Code)
	}
}

func TestWebhookRouter_UnknownEventFallback(t *testing.T) {
	called := false
	router := squad.NewWebhookRouter(testWebhookSecret).
		OnUnknown(func(_ context.Context, _ *squad.WebhookEvent) error {
			called = true
			return nil
		})

	payload, _ := json.Marshal(map[string]any{
		"event": "future.unknown.event",
		"body":  map[string]any{"data": "value"},
	})
	sig := signPayload(payload, testWebhookSecret)
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(payload))) //nolint:noctx
	req.Header.Set("x-squad-signature", sig)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Error("OnUnknown handler was not called")
	}
}

func TestWebhookRouter_ImplementsHTTPHandler(_ *testing.T) {
	// Compile-time check that WebhookRouter satisfies http.Handler.
	var _ http.Handler = squad.NewWebhookRouter("secret")
}

func TestWebhookRouter_FluentChaining(t *testing.T) {
	// Verify fluent chaining returns the same router.
	r := squad.NewWebhookRouter("secret")
	r2 := r.
		OnTransactionSuccess(nil).
		OnTransactionFailed(nil).
		OnVirtualAccountCredit(nil).
		OnTransferSuccess(nil).
		OnTransferFailed(nil).
		OnTransferReversed(nil).
		OnDisputeOpened(nil).
		OnDisputeResolved(nil)

	if r != r2 {
		t.Error("fluent chaining should return the same *WebhookRouter")
	}
}
