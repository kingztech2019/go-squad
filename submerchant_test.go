package squad_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kingztech2019/go-squad"
)

func TestSubMerchant_Create_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/merchant/sub-merchant" {
			t.Errorf("expected /merchant/sub-merchant, got %s", r.URL.Path)
		}
		writeJSON(w, 200, "success", squad.SubMerchant{
			ID:            "sub_001",
			MerchantID:    "merch_abc",
			DisplayName:   "Vendor Store",
			AccountName:   "Emeka Obi",
			AccountNumber: "0123456789",
			BankCode:      "057",
			BankName:      "Zenith Bank",
			Email:         "emeka@vendor.ng",
			Status:        "active",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.SubMerchants.Create(context.Background(), &squad.CreateSubMerchantParams{
		DisplayName:   "Vendor Store",
		AccountName:   "Emeka Obi",
		AccountNumber: "0123456789",
		BankCode:      "057",
		Email:         "emeka@vendor.ng",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "sub_001" {
		t.Errorf("expected sub_001, got %s", resp.ID)
	}
	if resp.Status != "active" {
		t.Errorf("expected active, got %s", resp.Status)
	}
}

func TestSubMerchant_Get_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		writeJSON(w, 200, "success", squad.SubMerchant{
			ID:          "sub_001",
			DisplayName: "Vendor Store",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.SubMerchants.Get(context.Background(), "sub_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DisplayName != "Vendor Store" {
		t.Errorf("expected Vendor Store, got %s", resp.DisplayName)
	}
}

func TestSubMerchant_List_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, "success", squad.SubMerchantListResponse{
			Merchants: []squad.SubMerchant{
				{ID: "sub_001", DisplayName: "Vendor A"},
				{ID: "sub_002", DisplayName: "Vendor B"},
			},
			Total:   2,
			Page:    1,
			PerPage: 20,
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.SubMerchants.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Merchants) != 2 {
		t.Errorf("expected 2 merchants, got %d", len(resp.Merchants))
	}
}

func TestSubMerchant_EmptyMerchantID(t *testing.T) {
	client := squad.New("sandbox_sk_test")
	_, err := client.SubMerchants.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty merchantID")
	}
	_, err = client.SubMerchants.Delete(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty merchantID on delete")
	}
}
