package squad_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kingztech2019/go-squad"
)

func TestFundsTransfer_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/payout/transfer" {
			t.Errorf("expected /payout/transfer, got %s", r.URL.Path)
		}
		writeJSON(w, 200, "success", squad.TransferResponse{
			TransactionRef: "pay_ref_001",
			Amount:         500000,
			Status:         "Success",
			AccountName:    "John Doe",
			AccountNumber:  "0123456789",
			BankName:       "Zenith Bank",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Transfers.FundsTransfer(context.Background(), &squad.FundsTransferParams{
		TransactionRef: "pay_ref_001",
		Amount:         500000,
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

func TestAccountLookup_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("bank_code") != "057" {
			t.Errorf("expected bank_code 057, got %s", r.URL.Query().Get("bank_code"))
		}
		if r.URL.Query().Get("account_number") != "0123456789" {
			t.Errorf("expected account_number 0123456789, got %s", r.URL.Query().Get("account_number"))
		}
		writeJSON(w, 200, "success", squad.AccountLookupResponse{
			AccountName:   "John Doe",
			AccountNumber: "0123456789",
			BankCode:      "057",
			BankName:      "Zenith Bank",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Transfers.AccountLookup(context.Background(), "057", "0123456789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccountName != "John Doe" {
		t.Errorf("expected John Doe, got %s", resp.AccountName)
	}
}

func TestAccountLookup_EmptyParams(t *testing.T) {
	client := squad.New("sandbox_sk_test")
	_, err := client.Transfers.AccountLookup(context.Background(), "", "0123456789")
	if err == nil {
		t.Fatal("expected error for empty bankCode")
	}
}

func TestGetAllTransfers_Pagination(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("per_page") != "10" {
			t.Errorf("expected per_page=10, got %s", r.URL.Query().Get("per_page"))
		}
		writeJSON(w, 200, "success", squad.TransferListResponse{
			Transfers: []squad.TransferStatusResponse{},
			Total:     0,
			Page:      2,
			PerPage:   10,
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Transfers.GetAllTransactions(context.Background(), &squad.TransferListParams{
		Page:    2,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Page != 2 {
		t.Errorf("expected page 2, got %d", resp.Page)
	}
}
