// Package squad provides a Go client for the Squad by GTCO payment gateway API.
//
// # Getting Started
//
//	client := squad.New("sandbox_sk_your_key_here")
//
//	resp, err := client.Transactions.InitiatePayment(ctx, &squad.InitiatePaymentParams{
//	    Email:    "customer@example.com",
//	    Amount:   squad.NGN(5000), // ₦5,000
//	    Currency: "NGN",
//	    CallbackURL: "https://yoursite.com/callback",
//	})
//
// # Environments
//
// The sandbox environment is automatically selected when your key starts with "sandbox_sk_".
// For production, use your live key or explicitly call WithProduction().
//
// # Idempotency
//
//	key, _ := squad.GenerateIdempotencyKey()
//	ctx = squad.WithIdempotencyKey(ctx, key)
//	resp, err := client.Transactions.InitiatePayment(ctx, params)
//
// # Pagination
//
//	iter := client.Transfers.All(ctx, nil)
//	for iter.Next() {
//	    fmt.Println(iter.Item().TransactionRef)
//	}
//
// # Webhook Verification
//
//	event, err := squad.ParseWebhook(body, r.Header.Get("x-squad-signature"), secretKey)
//
// # Webhook Router
//
//	router := squad.NewWebhookRouter(secretKey).
//	    OnTransactionSuccess(handlePayment).
//	    OnVirtualAccountCredit(handleCredit)
//	http.Handle("/webhook/squad", router)
package squad

import (
	"net/http"
	"strings"
	"time"
)

const (
	// Version is the current SDK version.
	Version = "1.1.0"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second

	// SandboxBaseURL is the Squad sandbox API base URL.
	SandboxBaseURL = "https://sandbox-api-d.squadco.com"

	// ProductionBaseURL is the Squad production API base URL.
	ProductionBaseURL = "https://api-d.squadco.com"
)

// Client is the root SDK object. All API services are accessible as fields.
// A Client is safe for concurrent use across goroutines.
type Client struct {
	secretKey       string
	baseURL         string
	httpClient      *http.Client
	userAgent       string
	logger          Logger
	beforeRequest   func(*http.Request)
	afterResponse   func(*http.Request, *http.Response, time.Duration)
	autoIdempotency bool

	// Transactions handles payment initiation, verification, and refunds.
	Transactions *TransactionService

	// VirtualAccounts handles NUBAN virtual account creation and management.
	VirtualAccounts *VirtualAccountService

	// Transfers handles fund transfers to Nigerian bank accounts.
	Transfers *TransferService

	// Disputes handles chargeback disputes and evidence submission.
	Disputes *DisputeService

	// VAS handles value-added services: airtime, data, cable TV, electricity, and SMS.
	VAS *VASService

	// SubMerchants manages sub-merchant accounts for aggregators and platforms.
	SubMerchants *SubMerchantService
}

// New constructs a ready-to-use Client.
//
// secretKey is your Squad secret key (sandbox_sk_* for sandbox, live secret for production).
// The sandbox environment is automatically selected when the key starts with "sandbox_sk_".
// Pass functional options to override defaults.
//
//	client := squad.New("sandbox_sk_xxx")
//	client := squad.New("live_sk_xxx", squad.WithLogger(squad.StdLogger()))
func New(secretKey string, opts ...Option) *Client {
	cfg := &config{
		baseURL:   ProductionBaseURL,
		timeout:   DefaultTimeout,
		userAgent: "go-squad/" + Version,
		logger:    noopLogger{},
	}

	// Auto-detect sandbox from key prefix before applying options.
	if strings.HasPrefix(secretKey, "sandbox_sk_") {
		cfg.baseURL = SandboxBaseURL
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: cfg.timeout}
	}

	if cfg.userAgent != "go-squad/"+Version {
		cfg.userAgent = "go-squad/" + Version + " " + cfg.userAgent
	}

	c := &Client{
		secretKey:       secretKey,
		baseURL:         cfg.baseURL,
		httpClient:      cfg.httpClient,
		userAgent:       cfg.userAgent,
		logger:          cfg.logger,
		beforeRequest:   cfg.beforeRequest,
		afterResponse:   cfg.afterResponse,
		autoIdempotency: cfg.autoIdempotency,
	}

	c.Transactions = &TransactionService{client: c}
	c.VirtualAccounts = &VirtualAccountService{client: c}
	c.Transfers = &TransferService{client: c}
	c.Disputes = &DisputeService{client: c}
	c.VAS = &VASService{client: c}
	c.SubMerchants = &SubMerchantService{client: c}

	return c
}
