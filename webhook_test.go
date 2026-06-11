package squad_test

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kingztech2019/go-squad"
)

const testWebhookSecret = "sandbox_sk_test_secret"

func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestParseWebhook_ValidSignature(t *testing.T) {
	payload := []byte(`{"event":"charge.success","body":{"transaction_ref":"txn_001","amount":500000,"currency":"NGN","transaction_status":"Success","channel":"card","customer_email":"user@example.com","customer_name":"John Doe","gateway_ref":"gw_001","is_recurring":false,"created_at":"2026-01-01T00:00:00Z"}}`)
	sig := signPayload(payload, testWebhookSecret)

	event, err := squad.ParseWebhook(payload, sig, testWebhookSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Event != squad.EventTransactionSuccess {
		t.Errorf("expected charge.success, got %s", event.Event)
	}
}

func TestParseWebhook_InvalidSignature(t *testing.T) {
	payload := []byte(`{"event":"charge.success","body":{}}`)
	_, err := squad.ParseWebhook(payload, "invalid_signature", testWebhookSecret)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
	if !errors.Is(err, squad.ErrInvalidSignature) {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestParseWebhook_TamperedPayload(t *testing.T) {
	original := []byte(`{"event":"charge.success","body":{"amount":500000}}`)
	sig := signPayload(original, testWebhookSecret)

	tampered := []byte(`{"event":"charge.success","body":{"amount":9999999}}`)
	_, err := squad.ParseWebhook(tampered, sig, testWebhookSecret)
	if !errors.Is(err, squad.ErrInvalidSignature) {
		t.Errorf("expected ErrInvalidSignature for tampered payload, got %v", err)
	}
}

func TestParseWebhook_MalformedJSON(t *testing.T) {
	payload := []byte(`not valid json`)
	sig := signPayload(payload, testWebhookSecret)

	_, err := squad.ParseWebhook(payload, sig, testWebhookSecret)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if errors.Is(err, squad.ErrInvalidSignature) {
		t.Error("should not be ErrInvalidSignature for malformed JSON (signature was valid)")
	}
}

func TestVerifySignature_Direct(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	sig := signPayload(payload, testWebhookSecret)

	if !squad.VerifySignature(payload, sig, testWebhookSecret) {
		t.Error("expected VerifySignature to return true for valid signature")
	}
	if squad.VerifySignature(payload, "wrong", testWebhookSecret) {
		t.Error("expected VerifySignature to return false for wrong signature")
	}
}

func TestWebhookEvent_ParseBody_Transaction(t *testing.T) {
	body := squad.WebhookTransactionBody{
		TransactionRef: "txn_001",
		Amount:         500000,
		Currency:       "NGN",
		Status:         "Success",
	}
	bodyBytes, _ := json.Marshal(body)
	event := &squad.WebhookEvent{
		Event: squad.EventTransactionSuccess,
		Body:  bodyBytes,
	}

	parsed, err := event.ParseBody()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	txn, ok := parsed.(*squad.WebhookTransactionBody)
	if !ok {
		t.Fatalf("expected *WebhookTransactionBody, got %T", parsed)
	}
	if txn.TransactionRef != "txn_001" {
		t.Errorf("expected txn_001, got %s", txn.TransactionRef)
	}
}

func TestWebhookEvent_ParseBody_VirtualAccount(t *testing.T) {
	body := squad.WebhookVirtualAccountBody{
		VirtualAccountNumber: "0123456789",
		Amount:               100000,
		CustomerIdentifier:   "cust_001",
	}
	bodyBytes, _ := json.Marshal(body)
	event := &squad.WebhookEvent{
		Event: squad.EventVirtualAccountCredit,
		Body:  bodyBytes,
	}

	parsed, err := event.ParseBody()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	va, ok := parsed.(*squad.WebhookVirtualAccountBody)
	if !ok {
		t.Fatalf("expected *WebhookVirtualAccountBody, got %T", parsed)
	}
	if va.VirtualAccountNumber != "0123456789" {
		t.Errorf("expected 0123456789, got %s", va.VirtualAccountNumber)
	}
}

func TestWebhookEvent_ParseBody_Transfer(t *testing.T) {
	body := squad.WebhookTransferBody{
		TransactionRef: "transfer_001",
		Amount:         200000,
		Status:         "Success",
	}
	bodyBytes, _ := json.Marshal(body)

	for _, evType := range []squad.EventType{
		squad.EventTransferSuccess,
		squad.EventTransferFailed,
		squad.EventTransferReversed,
	} {
		event := &squad.WebhookEvent{Event: evType, Body: bodyBytes}
		parsed, err := event.ParseBody()
		if err != nil {
			t.Fatalf("event %s: unexpected error: %v", evType, err)
		}
		if _, ok := parsed.(*squad.WebhookTransferBody); !ok {
			t.Errorf("event %s: expected *WebhookTransferBody, got %T", evType, parsed)
		}
	}
}

func TestWebhookEvent_ParseBody_Dispute(t *testing.T) {
	body := squad.WebhookDisputeBody{
		TicketID: "ticket_001",
		Amount:   150000,
		Status:   "open",
	}
	bodyBytes, _ := json.Marshal(body)
	event := &squad.WebhookEvent{
		Event: squad.EventDisputeOpened,
		Body:  bodyBytes,
	}

	parsed, err := event.ParseBody()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := parsed.(*squad.WebhookDisputeBody)
	if !ok {
		t.Fatalf("expected *WebhookDisputeBody, got %T", parsed)
	}
	if d.TicketID != "ticket_001" {
		t.Errorf("expected ticket_001, got %s", d.TicketID)
	}
}

func TestWebhookEvent_ParseBody_UnknownEvent(t *testing.T) {
	event := &squad.WebhookEvent{
		Event: "unknown.future.event",
		Body:  []byte(`{"foo":"bar"}`),
	}
	parsed, err := event.ParseBody()
	if err != nil {
		t.Fatalf("unexpected error for unknown event: %v", err)
	}
	if parsed == nil {
		t.Error("expected non-nil result for unknown event")
	}
}
