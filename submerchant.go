package squad

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// SubMerchantService manages sub-merchant accounts for aggregators and marketplace platforms.
// Requires an aggregator-level Squad account with the appropriate permissions.
type SubMerchantService struct {
	client *Client
}

// Create onboards a new sub-merchant under the aggregator account.
// The sub-merchant receives their own merchant ID and dashboard access.
func (s *SubMerchantService) Create(ctx context.Context, params *CreateSubMerchantParams) (*SubMerchant, error) {
	var out SubMerchant
	if err := s.client.do(ctx, http.MethodPost, "/merchant/sub-merchant", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves the details of a sub-merchant by their merchant ID.
func (s *SubMerchantService) Get(ctx context.Context, merchantID string) (*SubMerchant, error) {
	if merchantID == "" {
		return nil, fmt.Errorf("squad: merchantID must not be empty")
	}
	var out SubMerchant
	if err := s.client.do(ctx, http.MethodGet, "/merchant/sub-merchant/"+merchantID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a paginated list of all sub-merchants under the aggregator account.
func (s *SubMerchantService) List(ctx context.Context, params *SubMerchantListParams) (*SubMerchantListResponse, error) {
	q := url.Values{}
	if params != nil {
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
		}
	}
	var out SubMerchantListResponse
	if err := s.client.doGet(ctx, "/merchant/sub-merchant", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a sub-merchant from the aggregator account.
func (s *SubMerchantService) Delete(ctx context.Context, merchantID string) (*DeleteSubMerchantResponse, error) {
	if merchantID == "" {
		return nil, fmt.Errorf("squad: merchantID must not be empty")
	}
	var out DeleteSubMerchantResponse
	if err := s.client.do(ctx, http.MethodDelete, "/merchant/sub-merchant/"+merchantID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// All returns a lazy iterator over all sub-merchants, fetching pages on demand.
//
//	iter := client.SubMerchants.All(ctx, nil)
//	for iter.Next() {
//	    fmt.Println(iter.Item().DisplayName)
//	}
func (s *SubMerchantService) All(ctx context.Context, params *SubMerchantListParams) *Iter[SubMerchant] {
	perPage := 20
	if params != nil && params.PerPage > 0 {
		perPage = params.PerPage
	}
	return newIter(ctx, func(ctx context.Context, page int) ([]SubMerchant, error) {
		result, err := s.List(ctx, &SubMerchantListParams{Page: page, PerPage: perPage})
		if err != nil {
			return nil, err
		}
		return result.Merchants, nil
	})
}
