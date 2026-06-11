package squad

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// TransferService handles fund transfers to Nigerian bank accounts and Squad wallets.
type TransferService struct {
	client *Client
}

// FundsTransfer transfers funds from the Squad merchant wallet to any Nigerian bank account.
// Use AccountLookup first to verify the destination account before transferring.
func (s *TransferService) FundsTransfer(ctx context.Context, params *FundsTransferParams) (*TransferResponse, error) {
	var out TransferResponse
	if err := s.client.do(ctx, http.MethodPost, "/payout/transfer", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IntraTransfer transfers funds between two Squad wallet holders.
func (s *TransferService) IntraTransfer(ctx context.Context, params *IntraTransferParams) (*TransferResponse, error) {
	var out TransferResponse
	if err := s.client.do(ctx, http.MethodPost, "/payout/intra-squad-transfer", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AccountLookup verifies a bank account number and returns the account holder's name.
// Always call this before FundsTransfer to validate the destination account.
func (s *TransferService) AccountLookup(ctx context.Context, bankCode, accountNumber string) (*AccountLookupResponse, error) {
	if bankCode == "" || accountNumber == "" {
		return nil, fmt.Errorf("squad: bankCode and accountNumber must not be empty")
	}
	q := url.Values{}
	q.Set("bank_code", bankCode)
	q.Set("account_number", accountNumber)
	var out AccountLookupResponse
	if err := s.client.doGet(ctx, "/payout/account/lookup", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTransactionStatus retrieves the current status of a previously initiated transfer.
func (s *TransferService) GetTransactionStatus(ctx context.Context, transactionRef string) (*TransferStatusResponse, error) {
	if transactionRef == "" {
		return nil, fmt.Errorf("squad: transactionRef must not be empty")
	}
	var out TransferStatusResponse
	if err := s.client.do(ctx, http.MethodGet, "/payout/transaction/"+transactionRef, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// All returns a lazy iterator over all transfer transactions, fetching pages on demand.
// Filters from params (StartDate, EndDate, Status) are preserved across pages.
//
//	iter := client.Transfers.All(ctx, &squad.TransferListParams{Status: "Success"})
//	for iter.Next() {
//	    fmt.Println(iter.Item().TransactionRef, squad.FromKobo(iter.Item().Amount))
//	}
func (s *TransferService) All(ctx context.Context, params *TransferListParams) *Iter[TransferStatusResponse] {
	perPage := 20
	if params != nil && params.PerPage > 0 {
		perPage = params.PerPage
	}
	return newIter(ctx, func(ctx context.Context, page int) ([]TransferStatusResponse, error) {
		p := &TransferListParams{Page: page, PerPage: perPage}
		if params != nil {
			p.StartDate = params.StartDate
			p.EndDate = params.EndDate
			p.Status = params.Status
		}
		result, err := s.GetAllTransactions(ctx, p)
		if err != nil {
			return nil, err
		}
		return result.Transfers, nil
	})
}

// GetAllTransactions returns a paginated list of all transfer transactions.
func (s *TransferService) GetAllTransactions(ctx context.Context, params *TransferListParams) (*TransferListResponse, error) {
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
	var out TransferListResponse
	if err := s.client.doGet(ctx, "/payout/list", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
