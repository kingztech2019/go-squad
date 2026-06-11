package squad

// FundsTransferParams holds parameters for transferring funds to a Nigerian bank account.
// Amount is in kobo (lowest denomination). Settlement is T+1 by default.
type FundsTransferParams struct {
	TransactionRef string `json:"transaction_ref"`
	Amount         int64  `json:"amount"`
	BankCode       string `json:"bank_code"`
	AccountNumber  string `json:"account_number"`
	AccountName    string `json:"account_name"`
	Currency       string `json:"currency_id"` // "NGN"
	Remark         string `json:"remark,omitempty"`
}

// IntraTransferParams holds parameters for a Squad-to-Squad wallet transfer.
type IntraTransferParams struct {
	TransactionRef     string `json:"transaction_ref"`
	Amount             int64  `json:"amount"`
	SenderIdentifier   string `json:"sender_identifier"`
	ReceiverIdentifier string `json:"receiver_identifier"`
	Narration          string `json:"narration,omitempty"`
}

// TransferResponse is the Data payload returned by both FundsTransfer and IntraTransfer.
type TransferResponse struct {
	TransactionRef string  `json:"transaction_reference"`
	Amount         int64   `json:"amount"`
	Fee            float64 `json:"transaction_charge"`
	Status         string  `json:"status"`
	AccountName    string  `json:"account_name"`
	AccountNumber  string  `json:"account_number"`
	BankCode       string  `json:"bank_code"`
	BankName       string  `json:"bank_name"`
	CreatedAt      string  `json:"created_at"`
}

// AccountLookupResponse is returned by AccountLookup after verifying a bank account.
type AccountLookupResponse struct {
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	BankName      string `json:"bank_name"`
}

// TransferStatusResponse is returned by GetTransactionStatus.
type TransferStatusResponse struct {
	TransactionRef  string  `json:"transaction_ref"`
	Status          string  `json:"status"`
	Amount          int64   `json:"amount"`
	Fee             float64 `json:"fee"`
	ResponseCode    string  `json:"response_code"`
	ResponseMessage string  `json:"response_message"`
	UpdatedAt       string  `json:"updated_at"`
}

// TransferListParams holds pagination and filter parameters for listing transfers.
type TransferListParams struct {
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Status    string `json:"status,omitempty"` // "Success", "Processing", "Failed"
}

// TransferListResponse holds a paginated list of transfer records.
type TransferListResponse struct {
	Transfers []TransferStatusResponse `json:"transfers"`
	Total     int                      `json:"total"`
	Page      int                      `json:"page"`
	PerPage   int                      `json:"per_page"`
}
