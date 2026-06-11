package squad

// InitiatePaymentParams holds all parameters for creating a new payment transaction.
// Amount must be in the lowest currency denomination (kobo for NGN: 100000 = ₦1,000).
type InitiatePaymentParams struct {
	Email               string         `json:"email"`
	Amount              int64          `json:"amount"`
	Currency            string         `json:"currency"`
	InitiatorCustomerID string         `json:"initiator_customer_id,omitempty"`
	TransactionRef      string         `json:"transaction_ref,omitempty"`
	CallbackURL         string         `json:"callback_url,omitempty"`
	PaymentChannels     []string       `json:"payment_channels,omitempty"`
	Metadata            map[string]any `json:"metadata,omitempty"`
	PassCharge          bool           `json:"pass_charge,omitempty"`
	CustomerName        string         `json:"customer_name,omitempty"`
	IsRecurring         bool           `json:"is_recurring,omitempty"`
	PlanCode            string         `json:"plan_code,omitempty"`
	ChargeToken         *ChargeToken   `json:"charge_token,omitempty"`
}

// ChargeToken holds card tokenization data used for recurring charges.
type ChargeToken struct {
	Token       string `json:"token"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

// InitiatePaymentResponse is the Data payload returned by InitiatePayment.
// Redirect the end-user to CheckoutURL to complete payment.
// CheckoutURL is computed by the SDK from TransactionRef — Squad does not return it directly.
type InitiatePaymentResponse struct {
	// CheckoutURL is set by the SDK. Redirect the customer here to complete payment.
	CheckoutURL string `json:"-"`

	TransactionRef     string       `json:"transaction_ref"`
	Amount             int64        `json:"transaction_amount"`
	Currency           string       `json:"currency"`
	CallbackURL        string       `json:"callback_url"`
	IsRecurring        bool         `json:"is_recurring"`
	AuthorizedChannels []string     `json:"authorized_channels"`
	MerchantInfo       MerchantInfo `json:"merchant_info"`
}

// MerchantInfo contains basic merchant identity returned inside payment responses.
type MerchantInfo struct {
	MerchantName string `json:"merchant_name"`
	MerchantID   string `json:"merchant_id"`
}

// CustomerInfo is embedded in transaction responses.
type CustomerInfo struct {
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone_number"`
	MerchantID    string `json:"merchant_id"`
}

// VerifyTransactionResponse is the Data payload returned by VerifyTransaction.
type VerifyTransactionResponse struct {
	TransactionRef      string         `json:"transaction_ref"`
	Amount              int64          `json:"transaction_amount"`
	Fee                 float64        `json:"fee"`
	MerchantAmount      int64          `json:"merchant_amount"`
	Status              string         `json:"transaction_status"`
	Currency            string         `json:"transaction_currency_id"`
	Email               string         `json:"email"`
	TransactionType     string         `json:"transaction_type"`
	MerchantName        string         `json:"merchant_name"`
	MerchantEmail       string         `json:"merchant_email"`
	Meta                map[string]any `json:"meta,omitempty"`
	IsRecurring         bool           `json:"is_recurring"`
	ChargeToken         *ChargeToken   `json:"charge_token,omitempty"`
	CreatedAt           string         `json:"created_at"`
	UpdatedAt           string         `json:"updated_at"`
}

// RefundTransactionParams holds parameters for initiating a transaction refund.
// RefundType must be "Full" or "Partial". For partial refunds, Amount is required.
type RefundTransactionParams struct {
	GatewayTransactionRef string `json:"gateway_transaction_ref"`
	TransactionRef        string `json:"transaction_ref"`
	RefundType            string `json:"refund_type"`
	ReasonForRefund       string `json:"reason_for_refund"`
	Amount                int64  `json:"refund_amount,omitempty"`
	RRN                   string `json:"rrn,omitempty"`
}

// RefundTransactionResponse is the Data payload returned by RefundTransaction.
type RefundTransactionResponse struct {
	GatewayRef     string `json:"gateway_ref"`
	TransactionRef string `json:"transaction_ref"`
	RefundStatus   string `json:"refund_status"`
	AmountRefunded int64  `json:"amount_refunded"`
}

// MissedWebhookParams holds query parameters for GetMissedWebhookTransactions.
type MissedWebhookParams struct {
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
	Action  string `json:"action,omitempty"`
}

// MissedWebhookResponse is the Data payload returned by GetMissedWebhookTransactions.
type MissedWebhookResponse struct {
	Transactions []VerifyTransactionResponse `json:"transactions"`
	Total        int                         `json:"total"`
	Page         int                         `json:"page"`
	PerPage      int                         `json:"per_page"`
}

// USSDBank describes a single bank supporting USSD payments.
type USSDBank struct {
	BankCode string `json:"bank_code"`
	BankName string `json:"bank_name"`
	USSD     string `json:"ussd"`
}

// USSDbanksResponse is the Data payload returned by GetUSSDbanks.
type USSDbanksResponse struct {
	Banks []USSDBank `json:"banks"`
}
