package squad

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// TransactionService handles payment initiation, verification, and refunds.
type TransactionService struct {
	client *Client
}

// InitiatePayment creates a new payment transaction and returns a checkout URL.
// Redirect the end-user to Response.CheckoutURL to complete payment via the Squad modal.
func (s *TransactionService) InitiatePayment(ctx context.Context, params *InitiatePaymentParams) (*InitiatePaymentResponse, error) {
	var out InitiatePaymentResponse
	if err := s.client.do(ctx, http.MethodPost, "/transaction/initiate", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyTransaction retrieves the current status of a transaction by its reference.
// Call this after the user returns from the checkout URL or inside a webhook handler.
func (s *TransactionService) VerifyTransaction(ctx context.Context, transactionRef string) (*VerifyTransactionResponse, error) {
	if transactionRef == "" {
		return nil, fmt.Errorf("squad: transactionRef must not be empty")
	}
	var out VerifyTransactionResponse
	if err := s.client.do(ctx, http.MethodGet, "/transaction/verify/"+transactionRef, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RefundTransaction initiates a full or partial refund for a completed transaction.
// Set params.RefundType to "Full" or "Partial". Partial refunds require params.Amount.
func (s *TransactionService) RefundTransaction(ctx context.Context, params *RefundTransactionParams) (*RefundTransactionResponse, error) {
	var out RefundTransactionResponse
	if err := s.client.do(ctx, http.MethodPost, "/transaction/refund", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AllMissedWebhooks returns a lazy iterator over all missed webhook transactions.
//
//	iter := client.Transactions.AllMissedWebhooks(ctx, nil)
//	for iter.Next() {
//	    tx := iter.Item()
//	    fmt.Println(tx.TransactionRef, tx.Status)
//	}
func (s *TransactionService) AllMissedWebhooks(ctx context.Context, params *MissedWebhookParams) *Iter[VerifyTransactionResponse] {
	perPage := 20
	if params != nil && params.PerPage > 0 {
		perPage = params.PerPage
	}
	return newIter(ctx, func(ctx context.Context, page int) ([]VerifyTransactionResponse, error) {
		p := &MissedWebhookParams{Page: page, PerPage: perPage}
		if params != nil {
			p.Action = params.Action
		}
		result, err := s.GetMissedWebhookTransactions(ctx, p)
		if err != nil {
			return nil, err
		}
		return result.Transactions, nil
	})
}

// GetMissedWebhookTransactions retrieves transactions whose webhooks were not delivered.
// Use for reconciliation. Delete processed entries to prevent re-delivery.
func (s *TransactionService) GetMissedWebhookTransactions(ctx context.Context, params *MissedWebhookParams) (*MissedWebhookResponse, error) {
	q := url.Values{}
	if params != nil {
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
		}
		if params.Action != "" {
			q.Set("action", params.Action)
		}
	}
	var out MissedWebhookResponse
	if err := s.client.doGet(ctx, "/transaction/webhook/missed", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUSSDbanks returns the list of banks that support USSD payment collection.
func (s *TransactionService) GetUSSDbanks(ctx context.Context) (*USSDbanksResponse, error) {
	var out USSDbanksResponse
	if err := s.client.do(ctx, http.MethodGet, "/ussd/banklist", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
