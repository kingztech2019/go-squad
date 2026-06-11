// Package squadtest provides a mock Squad API server for use in tests.
//
// It allows you to test your Squad integration without making real API calls,
// giving you full control over responses and the ability to assert on requests.
//
// # Basic Usage
//
//	func TestMyCheckout(t *testing.T) {
//	    srv := squadtest.NewServer(t)
//
//	    srv.OnInitiatePayment(func(p *squad.InitiatePaymentParams) (*squad.InitiatePaymentResponse, error) {
//	        return &squad.InitiatePaymentResponse{
//	            CheckoutURL:    "https://fake-checkout.squadco.com/abc",
//	            TransactionRef: p.TransactionRef,
//	        }, nil
//	    })
//
//	    // Pass srv.Client() to your application code instead of a real squad.Client.
//	    myService := checkout.NewService(srv.Client())
//	    url, err := myService.StartCheckout("customer@example.com", 500000)
//	    if err != nil { t.Fatal(err) }
//	    if url == "" { t.Error("expected checkout URL") }
//
//	    // Assert the request your code sent.
//	    if srv.LastRequest().URL.Path != "/transaction/initiate" {
//	        t.Error("unexpected path")
//	    }
//	}
package squadtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	squad "github.com/kingztech2019/go-squad"
)

// Server is a mock Squad API server.
// Register handlers using the On* methods, then inject srv.Client() into your code under test.
type Server struct {
	mu       sync.RWMutex
	srv      *httptest.Server
	t        testing.TB
	routes   []route
	requests []*http.Request
}

type route struct {
	method  string
	prefix  string
	handler http.HandlerFunc
}

// NewServer creates and starts a mock Squad API server.
// The server is automatically shut down when the test ends via t.Cleanup.
func NewServer(t testing.TB) *Server {
	t.Helper()
	s := &Server{t: t}
	s.srv = httptest.NewServer(http.HandlerFunc(s.handle))
	t.Cleanup(s.srv.Close)
	return s
}

// Client returns a squad.Client pre-configured to call this mock server.
// Pass this to your application code instead of a real production client.
func (s *Server) Client() *squad.Client {
	return squad.New("sandbox_sk_test_key", squad.WithBaseURL(s.srv.URL))
}

// URL returns the base URL of the mock server.
func (s *Server) URL() string {
	return s.srv.URL
}

// Requests returns all HTTP requests received by the server, in order.
func (s *Server) Requests() []*http.Request {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]*http.Request, len(s.requests))
	copy(cp, s.requests)
	return cp
}

// LastRequest returns the most recent HTTP request received by the server.
// Returns nil if no requests have been received.
func (s *Server) LastRequest() *http.Request {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.requests) == 0 {
		return nil
	}
	return s.requests[len(s.requests)-1]
}

// RequestCount returns the total number of requests received.
func (s *Server) RequestCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.requests)
}

// Reset clears all registered routes and recorded requests.
func (s *Server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes = nil
	s.requests = nil
}

// ---- Transaction handlers ----

// OnInitiatePayment registers a handler for POST /transaction/initiate.
func (s *Server) OnInitiatePayment(fn func(*squad.InitiatePaymentParams) (*squad.InitiatePaymentResponse, error)) *Server {
	return s.register("POST", "/transaction/initiate", func(w http.ResponseWriter, r *http.Request) {
		var params squad.InitiatePaymentParams
		decodeJSON(r, &params)
		resp, err := fn(&params)
		s.writeResult(w, resp, err)
	})
}

// OnVerifyTransaction registers a handler for GET /transaction/verify/{ref}.
// The transaction reference extracted from the path is passed to fn.
func (s *Server) OnVerifyTransaction(fn func(ref string) (*squad.VerifyTransactionResponse, error)) *Server {
	return s.register("GET", "/transaction/verify/", func(w http.ResponseWriter, r *http.Request) {
		ref := strings.TrimPrefix(r.URL.Path, "/transaction/verify/")
		resp, err := fn(ref)
		s.writeResult(w, resp, err)
	})
}

// OnRefundTransaction registers a handler for POST /transaction/refund.
func (s *Server) OnRefundTransaction(fn func(*squad.RefundTransactionParams) (*squad.RefundTransactionResponse, error)) *Server {
	return s.register("POST", "/transaction/refund", func(w http.ResponseWriter, r *http.Request) {
		var params squad.RefundTransactionParams
		decodeJSON(r, &params)
		resp, err := fn(&params)
		s.writeResult(w, resp, err)
	})
}

// ---- Virtual Account handlers ----

// OnCreateVirtualAccount registers a handler for POST /virtual-account.
func (s *Server) OnCreateVirtualAccount(fn func(*squad.CreateVirtualAccountParams) (*squad.VirtualAccount, error)) *Server {
	return s.register("POST", "/virtual-account", func(w http.ResponseWriter, r *http.Request) {
		var params squad.CreateVirtualAccountParams
		decodeJSON(r, &params)
		resp, err := fn(&params)
		s.writeResult(w, resp, err)
	})
}

// OnQueryVirtualAccount registers a handler for GET /virtual-account/{number}.
func (s *Server) OnQueryVirtualAccount(fn func(accountNumber string) (*squad.VirtualAccount, error)) *Server {
	return s.register("GET", "/virtual-account/", func(w http.ResponseWriter, r *http.Request) {
		number := strings.TrimPrefix(r.URL.Path, "/virtual-account/")
		// Skip customer transaction routes handled separately.
		if strings.HasPrefix(number, "customer/") {
			s.writeError(w, 404, "no handler registered for "+r.URL.Path)
			return
		}
		resp, err := fn(number)
		s.writeResult(w, resp, err)
	})
}

// ---- Transfer handlers ----

// OnFundsTransfer registers a handler for POST /payout/transfer.
func (s *Server) OnFundsTransfer(fn func(*squad.FundsTransferParams) (*squad.TransferResponse, error)) *Server {
	return s.register("POST", "/payout/transfer", func(w http.ResponseWriter, r *http.Request) {
		var params squad.FundsTransferParams
		decodeJSON(r, &params)
		resp, err := fn(&params)
		s.writeResult(w, resp, err)
	})
}

// OnAccountLookup registers a handler for GET /payout/account/lookup.
// bankCode and accountNumber are extracted from query parameters.
func (s *Server) OnAccountLookup(fn func(bankCode, accountNumber string) (*squad.AccountLookupResponse, error)) *Server {
	return s.register("GET", "/payout/account/lookup", func(w http.ResponseWriter, r *http.Request) {
		bankCode := r.URL.Query().Get("bank_code")
		accountNumber := r.URL.Query().Get("account_number")
		resp, err := fn(bankCode, accountNumber)
		s.writeResult(w, resp, err)
	})
}

// ---- VAS handlers ----

// OnBuyAirtime registers a handler for POST /vas/airtime.
func (s *Server) OnBuyAirtime(fn func(*squad.BuyAirtimeParams) (*squad.VASTransactionResponse, error)) *Server {
	return s.register("POST", "/vas/airtime", func(w http.ResponseWriter, r *http.Request) {
		var params squad.BuyAirtimeParams
		decodeJSON(r, &params)
		resp, err := fn(&params)
		s.writeResult(w, resp, err)
	})
}

// OnBuyElectricity registers a handler for POST /vas/electricity.
func (s *Server) OnBuyElectricity(fn func(*squad.BuyElectricityParams) (*squad.ElectricityResponse, error)) *Server {
	return s.register("POST", "/vas/electricity", func(w http.ResponseWriter, r *http.Request) {
		var params squad.BuyElectricityParams
		decodeJSON(r, &params)
		resp, err := fn(&params)
		s.writeResult(w, resp, err)
	})
}

// ---- Sub-merchant handlers ----

// OnCreateSubMerchant registers a handler for POST /merchant/sub-merchant.
func (s *Server) OnCreateSubMerchant(fn func(*squad.CreateSubMerchantParams) (*squad.SubMerchant, error)) *Server {
	return s.register("POST", "/merchant/sub-merchant", func(w http.ResponseWriter, r *http.Request) {
		var params squad.CreateSubMerchantParams
		decodeJSON(r, &params)
		resp, err := fn(&params)
		s.writeResult(w, resp, err)
	})
}

// ---- Dispute handlers ----

// OnUploadEvidence registers a handler for POST /dispute/upload-evidence/{ticketID}.
func (s *Server) OnUploadEvidence(fn func(ticketID string) (*squad.EvidenceUploadResponse, error)) *Server {
	return s.register("POST", "/dispute/upload-evidence/", func(w http.ResponseWriter, r *http.Request) {
		ticketID := strings.TrimPrefix(r.URL.Path, "/dispute/upload-evidence/")
		resp, err := fn(ticketID)
		s.writeResult(w, resp, err)
	})
}

// ---- Custom handler ----

// Handle registers a fully custom http.HandlerFunc for the given method and path prefix.
// Use this for endpoints not covered by the On* convenience methods.
//
//	srv.Handle("GET", "/ussd/banklist", func(w http.ResponseWriter, r *http.Request) {
//	    squadtest.WriteJSON(w, 200, "success", squad.USSDbanksResponse{...})
//	})
func (s *Server) Handle(method, pathPrefix string, fn http.HandlerFunc) *Server {
	return s.register(method, pathPrefix, fn)
}

// ---- Helpers ----

// WriteJSON writes a Squad-envelope JSON response. Use inside custom Handle handlers.
func WriteJSON(w http.ResponseWriter, squadStatus int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  squadStatus,
		"message": message,
		"data":    data,
	})
}

// WriteError writes a Squad-envelope error JSON response. Use inside custom Handle handlers.
func WriteError(w http.ResponseWriter, httpStatus, squadStatus int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  squadStatus,
		"message": message,
		"data":    nil,
	})
}

// ---- internal ----

func (s *Server) register(method, prefix string, fn http.HandlerFunc) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Prepend so the most recently registered handler wins.
	s.routes = append([]route{{method: method, prefix: prefix, handler: fn}}, s.routes...)
	return s
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.requests = append(s.requests, r)
	routes := s.routes
	s.mu.Unlock()

	for _, rt := range routes {
		if rt.method == r.Method && strings.HasPrefix(r.URL.Path, rt.prefix) {
			rt.handler(w, r)
			return
		}
	}
	s.writeError(w, 404, fmt.Sprintf("squadtest: no handler for %s %s", r.Method, r.URL.Path))
}

func (s *Server) writeResult(w http.ResponseWriter, data any, err error) {
	if err != nil {
		s.writeError(w, 400, err.Error())
		return
	}
	WriteJSON(w, 200, "success", data)
}

func (s *Server) writeError(w http.ResponseWriter, httpStatus int, message string) {
	WriteError(w, httpStatus, httpStatus, message)
}

func decodeJSON(r *http.Request, out any) {
	json.NewDecoder(r.Body).Decode(out) //nolint:errcheck
}
