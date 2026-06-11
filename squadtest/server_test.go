package squadtest_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	squad "github.com/kingztech2019/go-squad"
	"github.com/kingztech2019/go-squad/squadtest"
)

func TestServer_OnInitiatePayment(t *testing.T) {
	srv := squadtest.NewServer(t)

	srv.OnInitiatePayment(func(p *squad.InitiatePaymentParams) (*squad.InitiatePaymentResponse, error) {
		if p.Email != "customer@example.com" {
			t.Errorf("expected customer@example.com, got %s", p.Email)
		}
		return &squad.InitiatePaymentResponse{
			CheckoutURL:    "https://fake-checkout.squadco.com/abc",
			TransactionRef: p.TransactionRef,
			Currency:       "NGN",
		}, nil
	})

	client := srv.Client()
	resp, err := client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{
		Email:          "customer@example.com",
		Amount:         squad.NGN(5000),
		Currency:       "NGN",
		TransactionRef: "test-ref-001",
		CallbackURL:    "https://example.com/callback",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CheckoutURL != "https://fake-checkout.squadco.com/abc" {
		t.Errorf("unexpected checkout URL: %s", resp.CheckoutURL)
	}
	if resp.TransactionRef != "test-ref-001" {
		t.Errorf("expected test-ref-001, got %s", resp.TransactionRef)
	}
}

func TestServer_OnVerifyTransaction(t *testing.T) {
	srv := squadtest.NewServer(t)

	srv.OnVerifyTransaction(func(ref string) (*squad.VerifyTransactionResponse, error) {
		if ref != "txn_abc" {
			t.Errorf("expected txn_abc, got %s", ref)
		}
		return &squad.VerifyTransactionResponse{
			TransactionRef: ref,
			Status:         "Success",
			Amount:         squad.NGN(5000),
		}, nil
	})

	client := srv.Client()
	txn, err := client.Transactions.VerifyTransaction(context.Background(), "txn_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txn.Status != "Success" {
		t.Errorf("expected Success, got %s", txn.Status)
	}
}

func TestServer_HandlerReturnsError(t *testing.T) {
	srv := squadtest.NewServer(t)

	srv.OnInitiatePayment(func(_ *squad.InitiatePaymentParams) (*squad.InitiatePaymentResponse, error) {
		return nil, errors.New("The email field is required")
	})

	client := srv.Client()
	_, err := client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{
		Amount:   squad.NGN(100),
		Currency: "NGN",
	})
	if err == nil {
		t.Fatal("expected error from handler")
	}
	if !squad.IsBadRequest(err) {
		t.Errorf("expected bad request error, got %v", err)
	}
}

func TestServer_OnFundsTransfer(t *testing.T) {
	srv := squadtest.NewServer(t)

	srv.OnFundsTransfer(func(p *squad.FundsTransferParams) (*squad.TransferResponse, error) {
		return &squad.TransferResponse{
			TransactionRef: p.TransactionRef,
			Amount:         p.Amount,
			Status:         "Success",
			AccountName:    p.AccountName,
		}, nil
	})

	client := srv.Client()
	resp, err := client.Transfers.FundsTransfer(context.Background(), &squad.FundsTransferParams{
		TransactionRef: "payout-001",
		Amount:         squad.NGN(2000),
		BankCode:       "057",
		AccountNumber:  "0123456789",
		AccountName:    "John Doe",
		Currency:       "NGN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "Success" {
		t.Errorf("expected Success, got %s", resp.Status)
	}
}

func TestServer_OnCreateSubMerchant(t *testing.T) {
	srv := squadtest.NewServer(t)

	srv.OnCreateSubMerchant(func(p *squad.CreateSubMerchantParams) (*squad.SubMerchant, error) {
		return &squad.SubMerchant{
			ID:          "sub_new",
			DisplayName: p.DisplayName,
			Status:      "active",
		}, nil
	})

	client := srv.Client()
	resp, err := client.SubMerchants.Create(context.Background(), &squad.CreateSubMerchantParams{
		DisplayName:   "My Marketplace Vendor",
		AccountName:   "Vendor Name",
		AccountNumber: "1234567890",
		BankCode:      "011",
		Email:         "vendor@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DisplayName != "My Marketplace Vendor" {
		t.Errorf("unexpected display name: %s", resp.DisplayName)
	}
}

func TestServer_RequestRecording(t *testing.T) {
	srv := squadtest.NewServer(t)

	srv.OnInitiatePayment(func(_ *squad.InitiatePaymentParams) (*squad.InitiatePaymentResponse, error) {
		return &squad.InitiatePaymentResponse{CheckoutURL: "https://x.com"}, nil
	})

	if srv.RequestCount() != 0 {
		t.Error("expected 0 requests initially")
	}

	client := srv.Client()
	client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{ //nolint:errcheck
		Email: "a@b.com", Amount: 100, Currency: "NGN",
	})
	client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{ //nolint:errcheck
		Email: "c@d.com", Amount: 200, Currency: "NGN",
	})

	if srv.RequestCount() != 2 {
		t.Errorf("expected 2 requests, got %d", srv.RequestCount())
	}
	if srv.LastRequest() == nil {
		t.Error("expected non-nil last request")
	}
}

func TestServer_Reset(t *testing.T) {
	srv := squadtest.NewServer(t)

	srv.OnInitiatePayment(func(_ *squad.InitiatePaymentParams) (*squad.InitiatePaymentResponse, error) {
		return &squad.InitiatePaymentResponse{CheckoutURL: "https://x.com"}, nil
	})

	client := srv.Client()
	client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{ //nolint:errcheck
		Email: "a@b.com", Amount: 100, Currency: "NGN",
	})

	srv.Reset()

	if srv.RequestCount() != 0 {
		t.Errorf("expected 0 requests after reset, got %d", srv.RequestCount())
	}

	// After reset, request should get 404.
	_, err := client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{
		Email: "a@b.com", Amount: 100, Currency: "NGN",
	})
	if err == nil {
		t.Error("expected error after reset (no handlers)")
	}
}

func TestServer_CustomHandle(t *testing.T) {
	srv := squadtest.NewServer(t)

	srv.Handle("GET", "/ussd/banklist", func(w http.ResponseWriter, _ *http.Request) {
		squadtest.WriteJSON(w, 200, "success", squad.USSDbanksResponse{
			Banks: []squad.USSDBank{
				{BankCode: "057", BankName: "Zenith Bank", USSD: "*966#"},
			},
		})
	})

	client := srv.Client()
	resp, err := client.Transactions.GetUSSDbanks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Banks) != 1 {
		t.Errorf("expected 1 bank, got %d", len(resp.Banks))
	}
}
