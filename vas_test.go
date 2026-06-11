package squad_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kingztech2019/go-squad"
)

func TestBuyAirtime_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/vas/airtime" {
			t.Errorf("expected /vas/airtime, got %s", r.URL.Path)
		}
		writeJSON(w, 200, "success", squad.VASTransactionResponse{
			TransactionRef: "vas_ref_001",
			Amount:         5000,
			Status:         "successful",
			PhoneNumber:    "2348012345678",
			Network:        "MTN",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.VAS.BuyAirtime(context.Background(), &squad.BuyAirtimeParams{
		PhoneNumber:    "2348012345678",
		Amount:         5000,
		Network:        "MTN",
		TransactionRef: "vas_ref_001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "successful" {
		t.Errorf("expected successful, got %s", resp.Status)
	}
}

func TestGetDataPlans_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vas/data-plans/MTN" {
			t.Errorf("expected /vas/data-plans/MTN, got %s", r.URL.Path)
		}
		writeJSON(w, 200, "success", squad.DataPlansResponse{
			Plans: []squad.DataPlan{
				{PlanCode: "mtn_1gb_30", PlanName: "1GB for 30 days", Amount: 300, Validity: "30 days", Network: "MTN"},
				{PlanCode: "mtn_2gb_30", PlanName: "2GB for 30 days", Amount: 500, Validity: "30 days", Network: "MTN"},
			},
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.VAS.GetDataPlans(context.Background(), "MTN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Plans) != 2 {
		t.Errorf("expected 2 plans, got %d", len(resp.Plans))
	}
}

func TestBuyElectricity_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, "success", squad.ElectricityResponse{
			VASTransactionResponse: squad.VASTransactionResponse{
				TransactionRef: "elec_ref_001",
				Amount:         500000,
				Status:         "successful",
			},
			MeterNumber:      "04123456789",
			Units:            "45.2",
			ElectricityToken: "1234-5678-9012-3456-7890",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.VAS.BuyElectricity(context.Background(), &squad.BuyElectricityParams{
		MeterNumber:    "04123456789",
		Amount:         500000,
		BillerCode:     "IKEDC",
		MeterType:      "prepaid",
		TransactionRef: "elec_ref_001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ElectricityToken == "" {
		t.Error("expected non-empty ElectricityToken")
	}
}

func TestGetElectricityBillers_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, "success", squad.ElectricityBillersResponse{
			Billers: []squad.ElectricityBiller{
				{BillerCode: "IKEDC", BillerName: "Ikeja Electric", MeterTypes: []string{"prepaid", "postpaid"}},
				{BillerCode: "AEDC", BillerName: "Abuja Electricity", MeterTypes: []string{"prepaid", "postpaid"}},
			},
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.VAS.GetElectricityBillers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Billers) != 2 {
		t.Errorf("expected 2 billers, got %d", len(resp.Billers))
	}
}

func TestSendSMS_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, "success", squad.SMSResponse{
			TransactionRef: "sms_ref_001",
			Status:         "sent",
			Recipients:     []string{"2348012345678"},
			MessageID:      "msg_abc123",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.VAS.SendSMS(context.Background(), &squad.SendSMSParams{
		To:             []string{"2348012345678"},
		From:           "MyBrand",
		Body:           "Your order has been confirmed.",
		TransactionRef: "sms_ref_001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "sent" {
		t.Errorf("expected sent, got %s", resp.Status)
	}
}
