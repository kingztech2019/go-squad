package squad

// CreateVirtualAccountParams holds all parameters to create a NUBAN virtual account.
// CBN profiling on the Squad dashboard is required before this endpoint is available.
type CreateVirtualAccountParams struct {
	CustomerIdentifier string `json:"customer_identifier"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	MiddleName         string `json:"middle_name,omitempty"`
	MobileNum          string `json:"mobile_num"`
	Email              string `json:"email"`
	BVN                string `json:"bvn"`
	DOB                string `json:"dob"` // "DD/MM/YYYY"
	Address            string `json:"address,omitempty"`
	Gender             string `json:"gender,omitempty"` // "1" = male, "2" = female
	BeneficiaryAccount string `json:"beneficiary_account,omitempty"`
}

// VirtualAccount is the core virtual account object returned by the API.
type VirtualAccount struct {
	VirtualAccountNumber string `json:"virtual_account_number"`
	CustomerIdentifier   string `json:"customer_identifier"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	MiddleName           string `json:"middle_name"`
	MobileNum            string `json:"mobile_num"`
	Email                string `json:"email"`
	UniqueID             string `json:"unique_id"`
	BeneficiaryAccount   string `json:"beneficiary_account"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

// VirtualAccountTxParams holds query parameters for fetching virtual account transactions.
type VirtualAccountTxParams struct {
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	StartDate string `json:"start_date,omitempty"` // "YYYY-MM-DD"
	EndDate   string `json:"end_date,omitempty"`
	Action    string `json:"action,omitempty"`
}

// VirtualAccountTxResponse holds paginated transactions for a virtual account.
type VirtualAccountTxResponse struct {
	Transactions []VirtualAccountTransaction `json:"transactions"`
	Total        int                         `json:"total"`
	Page         int                         `json:"page"`
	PerPage      int                         `json:"per_page"`
}

// VirtualAccountTransaction is a single credit transaction on a virtual account.
type VirtualAccountTransaction struct {
	TransactionRef      string `json:"transaction_ref"`
	Amount              int64  `json:"amount"`
	Currency            string `json:"currency"`
	SenderName          string `json:"sender_name"`
	SenderBank          string `json:"sender_bank"`
	SenderAccountNumber string `json:"sender_account_number"`
	Status              string `json:"transaction_status"`
	CreatedAt           string `json:"created_at"`
}

// UpdateVirtualAccountParams holds fields that can be modified on an existing virtual account.
type UpdateVirtualAccountParams struct {
	CustomerIdentifier string `json:"customer_identifier"`
	BVN                string `json:"bvn,omitempty"`
	FirstName          string `json:"first_name,omitempty"`
	LastName           string `json:"last_name,omitempty"`
	MiddleName         string `json:"middle_name,omitempty"`
	MobileNum          string `json:"mobile_num,omitempty"`
}

// SimulateVirtualAccountParams holds parameters for a sandbox credit simulation.
type SimulateVirtualAccountParams struct {
	VirtualAccountNumber string  `json:"virtual_account_number"`
	Amount               float64 `json:"amount"`
}

// SimulateResponse is returned by Simulate.
type SimulateResponse struct {
	TransactionRef string `json:"transaction_ref"`
	Status         string `json:"status"`
}
