package squad

// DisputeListParams holds pagination and filter parameters for listing disputes.
type DisputeListParams struct {
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Status    string `json:"status,omitempty"` // "open", "closed", "pending"
}

// DisputeListResponse holds a paginated list of disputes.
type DisputeListResponse struct {
	Disputes []Dispute `json:"disputes"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
}

// Dispute represents a single chargeback or dispute record.
type Dispute struct {
	TicketID         string `json:"ticket_id"`
	TransactionRef   string `json:"transaction_ref"`
	Amount           int64  `json:"amount"`
	Currency         string `json:"currency"`
	Status           string `json:"dispute_status"`
	Reason           string `json:"dispute_reason"`
	CustomerEmail    string `json:"customer_email"`
	CustomerName     string `json:"customer_name"`
	DueDate          string `json:"due_date"`
	MerchantResponse string `json:"merchant_response"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// DisputeEvidence represents evidence attached to a dispute.
type DisputeEvidence struct {
	TicketID    string `json:"ticket_id"`
	EvidenceURL string `json:"evidence_url"`
	FileName    string `json:"file_name"`
	UploadedAt  string `json:"uploaded_at"`
}

// EvidenceUploadResponse is returned after a successful evidence upload.
type EvidenceUploadResponse struct {
	TicketID  string `json:"ticket_id"`
	FileName  string `json:"file_name"`
	UploadURL string `json:"upload_url"`
	Status    string `json:"status"`
}

// DisputeActionResponse is returned by AcceptDispute and RejectDispute.
type DisputeActionResponse struct {
	TicketID string `json:"ticket_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}
