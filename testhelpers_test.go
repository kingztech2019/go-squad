package squad_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kingztech2019/go-squad"
)

// newTestServer creates an httptest.Server using the given handler.
// The returned teardown func shuts the server down and should be deferred.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	return srv, srv.Close
}

// newTestClient creates a squad.Client pointed at the given test server URL.
func newTestClient(t *testing.T, serverURL string) *squad.Client {
	t.Helper()
	return squad.New("sandbox_sk_test_key", squad.WithBaseURL(serverURL))
}

// squadEnvelope is the Squad API response envelope used in test fixtures.
type squadEnvelope struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// writeJSON writes a successful Squad-envelope JSON response.
func writeJSON(w http.ResponseWriter, status int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(squadEnvelope{Status: status, Message: message, Data: data})
}

// writeErrorJSON writes an error Squad-envelope JSON response.
func writeErrorJSON(w http.ResponseWriter, httpStatus, squadStatus int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(squadEnvelope{Status: squadStatus, Message: message, Data: nil})
}
