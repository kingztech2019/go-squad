package squad_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kingztech2019/go-squad"
)

func TestVirtualAccount_Create_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/virtual-account" {
			t.Errorf("expected /virtual-account, got %s", r.URL.Path)
		}
		writeJSON(w, 200, "success", squad.VirtualAccount{
			VirtualAccountNumber: "0123456789",
			CustomerIdentifier:   "cust_001",
			FirstName:            "Adaeze",
			LastName:             "Okafor",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.VirtualAccounts.Create(context.Background(), &squad.CreateVirtualAccountParams{
		CustomerIdentifier: "cust_001",
		FirstName:          "Adaeze",
		LastName:           "Okafor",
		MobileNum:          "2348012345678",
		Email:              "adaeze@example.com",
		BVN:                "12345678901",
		DOB:                "01/01/1990",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.VirtualAccountNumber != "0123456789" {
		t.Errorf("expected 0123456789, got %s", resp.VirtualAccountNumber)
	}
}

func TestVirtualAccount_GetTransactions_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, "success", squad.VirtualAccountTxResponse{
			Transactions: []squad.VirtualAccountTransaction{
				{TransactionRef: "va_txn_001", Amount: 100000, Currency: "NGN"},
			},
			Total:   1,
			Page:    1,
			PerPage: 20,
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.VirtualAccounts.GetTransactions(context.Background(), "cust_001", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(resp.Transactions))
	}
}

func TestVirtualAccount_GetTransactions_EmptyIdentifier(t *testing.T) {
	client := squad.New("sandbox_sk_test")
	_, err := client.VirtualAccounts.GetTransactions(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty customerIdentifier")
	}
}

func TestVirtualAccount_Simulate_ProductionGuard(t *testing.T) {
	client := squad.New("live_sk_production_key", squad.WithBaseURL("https://api-d.squadco.com"))
	_, err := client.VirtualAccounts.Simulate(context.Background(), &squad.SimulateVirtualAccountParams{
		VirtualAccountNumber: "0123456789",
		Amount:               5000,
	})
	if err == nil {
		t.Fatal("expected error when calling Simulate in production")
	}
}
