package squad

import (
	"context"
	"net/http"
)

// VASService handles value-added services: airtime, data, cable TV, electricity, and SMS.
// All purchases are charged from the Squad merchant wallet.
type VASService struct {
	client *Client
}

// BuyAirtime purchases airtime and credits the specified phone number.
// Minimum purchase amount is 50 NGN. Supported networks: MTN, AIRTEL, GLO, 9MOBILE.
func (s *VASService) BuyAirtime(ctx context.Context, params *BuyAirtimeParams) (*VASTransactionResponse, error) {
	var out VASTransactionResponse
	if err := s.client.do(ctx, http.MethodPost, "/vas/airtime", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDataPlans retrieves available data bundle plans for the given network provider.
// networkProvider is one of: "MTN", "AIRTEL", "GLO", "9MOBILE".
func (s *VASService) GetDataPlans(ctx context.Context, networkProvider string) (*DataPlansResponse, error) {
	var out DataPlansResponse
	if err := s.client.do(ctx, http.MethodGet, "/vas/data-plans/"+networkProvider, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BuyData purchases a data bundle for a phone number.
// Use GetDataPlans to retrieve valid plan codes before calling this.
func (s *VASService) BuyData(ctx context.Context, params *BuyDataParams) (*VASTransactionResponse, error) {
	var out VASTransactionResponse
	if err := s.client.do(ctx, http.MethodPost, "/vas/data", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCablePackages retrieves available cable TV subscription packages for the given provider.
// cableProvider is one of: "DSTV", "GOTV", "STARTIMES".
func (s *VASService) GetCablePackages(ctx context.Context, cableProvider string) (*CablePackagesResponse, error) {
	var out CablePackagesResponse
	if err := s.client.do(ctx, http.MethodGet, "/vas/cable-packages/"+cableProvider, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BuyCable subscribes to a cable TV package for a smart card number.
// Use GetCablePackages to retrieve valid package codes before calling this.
func (s *VASService) BuyCable(ctx context.Context, params *BuyCableParams) (*VASTransactionResponse, error) {
	var out VASTransactionResponse
	if err := s.client.do(ctx, http.MethodPost, "/vas/cable", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetElectricityBillers retrieves the list of available electricity distribution companies (DISCOs).
func (s *VASService) GetElectricityBillers(ctx context.Context) (*ElectricityBillersResponse, error) {
	var out ElectricityBillersResponse
	if err := s.client.do(ctx, http.MethodGet, "/vas/electricity-billers", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BuyElectricity purchases electricity units for a prepaid or postpaid meter.
// The returned ElectricityToken must be entered into the customer's meter.
func (s *VASService) BuyElectricity(ctx context.Context, params *BuyElectricityParams) (*ElectricityResponse, error) {
	var out ElectricityResponse
	if err := s.client.do(ctx, http.MethodPost, "/vas/electricity", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendSMS sends a personalised SMS message to one or more recipients.
func (s *VASService) SendSMS(ctx context.Context, params *SendSMSParams) (*SMSResponse, error) {
	var out SMSResponse
	if err := s.client.do(ctx, http.MethodPost, "/vas/send-sms", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
