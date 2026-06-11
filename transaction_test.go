package squad_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/kingztech2019/go-squad"
)

func TestInitiatePayment_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/transaction/initiate" {
			t.Errorf("expected path /transaction/initiate, got %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Error("missing Authorization header")
		}
		writeJSON(w, 200, "success", squad.InitiatePaymentResponse{
			TransactionRef: "txn_test_001",
			Currency:       "NGN",
			Amount:         500000,
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{
		Email:          "customer@example.com",
		Amount:         500000,
		Currency:       "NGN",
		TransactionRef: "txn_test_001",
		CallbackURL:    "https://example.com/callback",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CheckoutURL == "" {
		t.Error("expected non-empty CheckoutURL")
	}
	if resp.TransactionRef != "txn_test_001" {
		t.Errorf("expected txn_test_001, got %s", resp.TransactionRef)
	}
}

func TestInitiatePayment_Unauthorized(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeErrorJSON(w, 401, 401, "Unauthorized")
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	_, err := client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{
		Email:    "customer@example.com",
		Amount:   500000,
		Currency: "NGN",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !squad.IsUnauthorized(err) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestInitiatePayment_BadRequest(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeErrorJSON(w, 400, 400, "The email field is required")
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	_, err := client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{
		Amount:   500000,
		Currency: "NGN",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !squad.IsBadRequest(err) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

func TestVerifyTransaction_Success(t *testing.T) {
	tests := []struct {
		name           string
		transactionRef string
		status         string
		wantErr        bool
	}{
		{name: "success", transactionRef: "txn_001", status: "Success"},
		{name: "pending", transactionRef: "txn_002", status: "Pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, tt.transactionRef) {
					t.Errorf("expected path to contain %s, got %s", tt.transactionRef, r.URL.Path)
				}
				writeJSON(w, 200, "success", squad.VerifyTransactionResponse{
					TransactionRef: tt.transactionRef,
					Status:         tt.status,
					Amount:         500000,
					Currency:       "NGN",
				})
			})
			defer teardown()

			client := newTestClient(t, srv.URL)
			resp, err := client.Transactions.VerifyTransaction(context.Background(), tt.transactionRef)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Status != tt.status {
				t.Errorf("expected status %s, got %s", tt.status, resp.Status)
			}
		})
	}
}

func TestVerifyTransaction_EmptyRef(t *testing.T) {
	client := squad.New("sandbox_sk_test")
	_, err := client.Transactions.VerifyTransaction(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty transactionRef")
	}
}

func TestRefundTransaction_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		writeJSON(w, 200, "success", squad.RefundTransactionResponse{
			TransactionRef: "txn_001",
			RefundStatus:   "processing",
			AmountRefunded: 500000,
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Transactions.RefundTransaction(context.Background(), &squad.RefundTransactionParams{
		GatewayTransactionRef: "gw_ref_001",
		TransactionRef:        "txn_001",
		RefundType:            "Full",
		ReasonForRefund:       "Customer requested refund",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RefundStatus == "" {
		t.Error("expected non-empty RefundStatus")
	}
}

func TestGetUSSDbanks_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, "success", squad.USSDbanksResponse{
			Banks: []squad.USSDBank{
				{BankCode: "057", BankName: "Zenith Bank", USSD: "*966#"},
				{BankCode: "011", BankName: "First Bank", USSD: "*894#"},
			},
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Transactions.GetUSSDbanks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Banks) != 2 {
		t.Errorf("expected 2 banks, got %d", len(resp.Banks))
	}
}
