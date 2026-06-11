package squad

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// DisputeService handles chargeback disputes and evidence submission.
type DisputeService struct {
	client *Client
}

// All returns a lazy iterator over all disputes, fetching pages on demand.
// Filters from params (Status, StartDate, EndDate) are preserved across pages.
//
//	iter := client.Disputes.All(ctx, &squad.DisputeListParams{Status: "open"})
//	for iter.Next() {
//	    d := iter.Item()
//	    fmt.Println(d.TicketID, d.Reason)
//	}
func (s *DisputeService) All(ctx context.Context, params *DisputeListParams) *Iter[Dispute] {
	perPage := 20
	if params != nil && params.PerPage > 0 {
		perPage = params.PerPage
	}
	return newIter(ctx, func(ctx context.Context, page int) ([]Dispute, error) {
		p := &DisputeListParams{Page: page, PerPage: perPage}
		if params != nil {
			p.StartDate = params.StartDate
			p.EndDate = params.EndDate
			p.Status = params.Status
		}
		result, err := s.GetAllDisputes(ctx, p)
		if err != nil {
			return nil, err
		}
		return result.Disputes, nil
	})
}

// GetAllDisputes retrieves a paginated list of all disputes on the merchant account.
func (s *DisputeService) GetAllDisputes(ctx context.Context, params *DisputeListParams) (*DisputeListResponse, error) {
	q := url.Values{}
	if params != nil {
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
		}
		if params.StartDate != "" {
			q.Set("start_date", params.StartDate)
		}
		if params.EndDate != "" {
			q.Set("end_date", params.EndDate)
		}
		if params.Status != "" {
			q.Set("status", params.Status)
		}
	}
	var out DisputeListResponse
	if err := s.client.doGet(ctx, "/dispute", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDisputeDetail retrieves the full details of a single dispute by its ticket ID.
func (s *DisputeService) GetDisputeDetail(ctx context.Context, ticketID string) (*Dispute, error) {
	if ticketID == "" {
		return nil, fmt.Errorf("squad: ticketID must not be empty")
	}
	var out Dispute
	if err := s.client.do(ctx, http.MethodGet, "/dispute/"+ticketID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDisputeEvidence retrieves any previously uploaded evidence for a dispute.
func (s *DisputeService) GetDisputeEvidence(ctx context.Context, ticketID string) (*DisputeEvidence, error) {
	if ticketID == "" {
		return nil, fmt.Errorf("squad: ticketID must not be empty")
	}
	var out DisputeEvidence
	if err := s.client.do(ctx, http.MethodGet, "/dispute/evidence/"+ticketID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadEvidence uploads a file as evidence for a dispute.
// fileData is the raw file bytes (PDF, PNG, or JPG). fileName includes the extension.
// Call this before RejectDispute — evidence must be present to contest a chargeback.
func (s *DisputeService) UploadEvidence(ctx context.Context, ticketID string, fileData []byte, fileName string) (*EvidenceUploadResponse, error) {
	if ticketID == "" {
		return nil, fmt.Errorf("squad: ticketID must not be empty")
	}
	if len(fileData) == 0 {
		return nil, fmt.Errorf("squad: fileData must not be empty")
	}
	fields := map[string]string{"ticket_id": ticketID}
	var out EvidenceUploadResponse
	if err := s.client.doMultipart(ctx, "/dispute/upload-evidence/"+ticketID, fields, "file", fileName, fileData, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AcceptDispute accepts a chargeback (merchant concedes the dispute).
func (s *DisputeService) AcceptDispute(ctx context.Context, ticketID string) (*DisputeActionResponse, error) {
	if ticketID == "" {
		return nil, fmt.Errorf("squad: ticketID must not be empty")
	}
	var out DisputeActionResponse
	if err := s.client.do(ctx, http.MethodPost, "/dispute/accept/"+ticketID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RejectDispute contests a chargeback (merchant disputes the claim).
// Evidence must be uploaded via UploadEvidence before calling this method.
func (s *DisputeService) RejectDispute(ctx context.Context, ticketID string) (*DisputeActionResponse, error) {
	if ticketID == "" {
		return nil, fmt.Errorf("squad: ticketID must not be empty")
	}
	var out DisputeActionResponse
	if err := s.client.do(ctx, http.MethodPost, "/dispute/reject/"+ticketID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
