package squad

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// VirtualAccountService handles NUBAN virtual account creation and management.
type VirtualAccountService struct {
	client *Client
}

// Create creates a NUBAN-compliant virtual account for a customer.
// The account is permanently tied to CustomerIdentifier and receives payments on their behalf.
// CBN profiling must be completed on the Squad dashboard before calling this endpoint.
func (s *VirtualAccountService) Create(ctx context.Context, params *CreateVirtualAccountParams) (*VirtualAccount, error) {
	var out VirtualAccount
	if err := s.client.do(ctx, http.MethodPost, "/virtual-account", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AllTransactions returns a lazy iterator over all transactions for a virtual account.
// Filters from params (StartDate, EndDate, Action) are preserved across pages.
//
//	iter := client.VirtualAccounts.AllTransactions(ctx, "cust-001", nil)
//	for iter.Next() {
//	    tx := iter.Item()
//	    fmt.Println(tx.TransactionRef, tx.SenderName)
//	}
func (s *VirtualAccountService) AllTransactions(ctx context.Context, customerIdentifier string, params *VirtualAccountTxParams) *Iter[VirtualAccountTransaction] {
	perPage := 20
	if params != nil && params.PerPage > 0 {
		perPage = params.PerPage
	}
	return newIter(ctx, func(ctx context.Context, page int) ([]VirtualAccountTransaction, error) {
		p := &VirtualAccountTxParams{Page: page, PerPage: perPage}
		if params != nil {
			p.StartDate = params.StartDate
			p.EndDate = params.EndDate
			p.Action = params.Action
		}
		result, err := s.GetTransactions(ctx, customerIdentifier, p)
		if err != nil {
			return nil, err
		}
		return result.Transactions, nil
	})
}

// GetTransactions retrieves paginated transactions for a virtual account by customer identifier.
func (s *VirtualAccountService) GetTransactions(ctx context.Context, customerIdentifier string, params *VirtualAccountTxParams) (*VirtualAccountTxResponse, error) {
	if customerIdentifier == "" {
		return nil, fmt.Errorf("squad: customerIdentifier must not be empty")
	}
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
		if params.Action != "" {
			q.Set("action", params.Action)
		}
	}
	var out VirtualAccountTxResponse
	if err := s.client.doGet(ctx, "/virtual-account/customer/"+customerIdentifier, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Query retrieves the details of a virtual account by its account number.
func (s *VirtualAccountService) Query(ctx context.Context, virtualAccountNumber string) (*VirtualAccount, error) {
	if virtualAccountNumber == "" {
		return nil, fmt.Errorf("squad: virtualAccountNumber must not be empty")
	}
	var out VirtualAccount
	if err := s.client.do(ctx, http.MethodGet, "/virtual-account/"+virtualAccountNumber, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies metadata on an existing virtual account.
func (s *VirtualAccountService) Update(ctx context.Context, params *UpdateVirtualAccountParams) (*VirtualAccount, error) {
	var out VirtualAccount
	if err := s.client.do(ctx, http.MethodPut, "/virtual-account", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Simulate credits a sandbox virtual account with a test transaction.
// This method returns an error when called against a production base URL.
func (s *VirtualAccountService) Simulate(ctx context.Context, params *SimulateVirtualAccountParams) (*SimulateResponse, error) {
	if !strings.Contains(s.client.baseURL, "sandbox") {
		return nil, fmt.Errorf("squad: Simulate is only available in sandbox mode")
	}
	var out SimulateResponse
	if err := s.client.do(ctx, http.MethodPost, "/virtual-account/simulate/credit", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
