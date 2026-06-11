package squad_test

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/kingztech2019/go-squad"
)

func TestGetAllDisputes_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, "success", squad.DisputeListResponse{
			Disputes: []squad.Dispute{
				{TicketID: "ticket_001", Amount: 200000, Status: "open"},
			},
			Total:   1,
			Page:    1,
			PerPage: 20,
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Disputes.GetAllDisputes(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Disputes) != 1 {
		t.Errorf("expected 1 dispute, got %d", len(resp.Disputes))
	}
}

func TestAcceptDispute_Success(t *testing.T) {
	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "accept") {
			t.Errorf("expected path to contain 'accept', got %s", r.URL.Path)
		}
		writeJSON(w, 200, "success", squad.DisputeActionResponse{
			TicketID: "ticket_001",
			Status:   "accepted",
			Message:  "Dispute accepted",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Disputes.AcceptDispute(context.Background(), "ticket_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "accepted" {
		t.Errorf("expected accepted, got %s", resp.Status)
	}
}

func TestUploadEvidence_MultipartEncoding(t *testing.T) {
	fileContent := []byte("fake PDF content for testing")

	srv, teardown := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(ct)
		if err != nil || mediaType != "multipart/form-data" {
			t.Errorf("expected multipart/form-data, got %s", ct)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		foundFile := false
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("multipart read error: %v", err)
			}
			if p.FormName() == "file" {
				foundFile = true
				data, _ := io.ReadAll(p)
				if string(data) != string(fileContent) {
					t.Error("file content mismatch")
				}
			}
		}
		if !foundFile {
			t.Error("file field not found in multipart request")
		}
		writeJSON(w, 200, "success", squad.EvidenceUploadResponse{
			TicketID: "ticket_001",
			Status:   "uploaded",
		})
	})
	defer teardown()

	client := newTestClient(t, srv.URL)
	resp, err := client.Disputes.UploadEvidence(context.Background(), "ticket_001", fileContent, "evidence.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "uploaded" {
		t.Errorf("expected uploaded, got %s", resp.Status)
	}
}

func TestUploadEvidence_EmptyFile(t *testing.T) {
	client := squad.New("sandbox_sk_test")
	_, err := client.Disputes.UploadEvidence(context.Background(), "ticket_001", []byte{}, "evidence.pdf")
	if err == nil {
		t.Fatal("expected error for empty fileData")
	}
}

func TestDisputeAction_EmptyTicketID(t *testing.T) {
	client := squad.New("sandbox_sk_test")
	_, err := client.Disputes.AcceptDispute(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty ticketID")
	}
	_, err = client.Disputes.RejectDispute(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty ticketID")
	}
}
