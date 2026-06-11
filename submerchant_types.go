package squad

// CreateSubMerchantParams holds parameters for onboarding a sub-merchant.
// Requires an aggregator-level Squad account.
type CreateSubMerchantParams struct {
	DisplayName   string `json:"display_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	Email         string `json:"email"`
	MobileNumber  string `json:"mobile_number,omitempty"`
}

// SubMerchant represents a sub-merchant account under an aggregator.
type SubMerchant struct {
	ID            string `json:"id"`
	MerchantID    string `json:"merchant_id"`
	DisplayName   string `json:"display_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	BankName      string `json:"bank_name"`
	Email         string `json:"email"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// SubMerchantListParams holds pagination parameters for listing sub-merchants.
type SubMerchantListParams struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

// SubMerchantListResponse holds a paginated list of sub-merchants.
type SubMerchantListResponse struct {
	Merchants []SubMerchant `json:"merchants"`
	Total     int           `json:"total"`
	Page      int           `json:"page"`
	PerPage   int           `json:"per_page"`
}

// DeleteSubMerchantResponse is returned when a sub-merchant is removed.
type DeleteSubMerchantResponse struct {
	MerchantID string `json:"merchant_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}
