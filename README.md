# go-squad

[![CI](https://github.com/kingztech2019/go-squad/actions/workflows/ci.yml/badge.svg)](https://github.com/kingztech2019/go-squad/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/kingztech2019/go-squad.svg)](https://pkg.go.dev/github.com/kingztech2019/go-squad)
[![Go Report Card](https://goreportcard.com/badge/github.com/kingztech2019/go-squad)](https://goreportcard.com/report/github.com/kingztech2019/go-squad)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

The comprehensive, idiomatic Go SDK for the [Squad by GTCO](https://squadco.com) payment gateway.

**What's included:**
- Payment initiation, verification, and refunds
- Virtual accounts (NUBAN-compliant)
- Fund transfers and account lookup
- Sub-merchant management for aggregators and marketplaces
- Dispute management with evidence upload
- Value-added services: airtime, data, cable TV, electricity, SMS
- Webhook signature validation + typed event router
- Auto-pagination iterator — loop over thousands of records without manual paging
- Idempotency keys — prevent duplicate charges on retried requests
- Structured logging interface — works with `log/slog`, `zap`, `zerolog`, and others
- Request/response hooks — for metrics, tracing, and custom headers
- `squadtest` package — mock Squad API server for unit testing your integration
- Zero external runtime dependencies
- Context-aware — every method accepts `context.Context`

---

## Installation

```bash
go get github.com/kingztech2019/go-squad
```

Requires Go 1.21 or later.

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    squad "github.com/kingztech2019/go-squad"
)

func main() {
    // Sandbox is auto-detected from the "sandbox_sk_" key prefix.
    client := squad.New(os.Getenv("SQUAD_SECRET_KEY"))

    resp, err := client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{
        Email:       "customer@example.com",
        Amount:      squad.NGN(5000), // ₦5,000 — no kobo confusion
        Currency:    "NGN",
        CallbackURL: "https://yoursite.com/callback",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Redirect customer to: %s\n", resp.CheckoutURL)
}
```

---

## Configuration

```go
// Sandbox auto-detected from key prefix
client := squad.New("sandbox_sk_xxx")

// Explicit environment
client := squad.New(key, squad.WithSandbox())
client := squad.New(key, squad.WithProduction())

// Development logging (stdout)
client := squad.New(key, squad.WithLogger(squad.StdLogger()))

// Production logging (your logger, e.g. slog)
type slogAdapter struct{ l *slog.Logger }
func (a slogAdapter) Info(msg string, kv ...any)  { a.l.Info(msg, kv...) }
func (a slogAdapter) Error(msg string, kv ...any) { a.l.Error(msg, kv...) }
client := squad.New(key, squad.WithLogger(slogAdapter{slog.Default()}))

// Metrics + distributed tracing hooks
client := squad.New(key,
    squad.WithBeforeRequest(func(req *http.Request) {
        req.Header.Set("X-Correlation-ID", getCorrelationID())
    }),
    squad.WithAfterResponse(func(req *http.Request, resp *http.Response, d time.Duration) {
        metrics.Record("squad.request", req.URL.Path, resp.StatusCode, d)
    }),
)

// Auto-generate idempotency keys on all POST requests
client := squad.New(key, squad.WithAutoIdempotency())

// Custom timeout
client := squad.New(key, squad.WithTimeout(60*time.Second))

// Custom HTTP client (retry, custom TLS, etc.)
client := squad.New(key, squad.WithHTTPClient(myHTTPClient))
```

---

## Money Helpers

The Squad API uses the lowest currency denomination (kobo for NGN, cents for USD). These helpers eliminate the most common payment integration bug:

```go
squad.NGN(5000)       // → 500000 (₦5,000 in kobo)
squad.NGN(1)          // → 100    (₦1.00 in kobo)
squad.USD(50)         // → 5000   ($50.00 in cents)
squad.FromKobo(50000) // → 500.0  (display value)
squad.FromCents(5000) // → 50.0   (display value)
```

---

## Idempotency Keys

Protect against duplicate charges when a request is retried after a network failure. Store the key **before** making the request so you can reuse it on retry.

```go
// Tie the key to your business operation
key, err := squad.GenerateIdempotencyKey()
if err != nil { log.Fatal(err) }

// Store key in your DB alongside the order before calling Squad.
ctx = squad.WithIdempotencyKey(ctx, "order-"+orderID+"-"+key)

resp, err := client.Transactions.InitiatePayment(ctx, params)
// If this times out and you retry, use the SAME ctx — same key, no double charge.
```

Or let the SDK auto-generate keys for every POST:
```go
client := squad.New(key, squad.WithAutoIdempotency())
```

---

## Transactions

### Initiate Payment

```go
resp, err := client.Transactions.InitiatePayment(ctx, &squad.InitiatePaymentParams{
    Email:           "customer@example.com",
    Amount:          squad.NGN(1000),
    Currency:        "NGN",
    TransactionRef:  "order-ref-001",
    CallbackURL:     "https://yoursite.com/callback",
    PaymentChannels: []string{"card", "bank", "ussd", "transfer"},
    CustomerName:    "John Doe",
    Metadata:        map[string]any{"order_id": "123"},
})
// Redirect user to resp.CheckoutURL
```

### Verify Transaction

```go
txn, err := client.Transactions.VerifyTransaction(ctx, "order-ref-001")
if txn.Status == "Success" {
    fmt.Printf("Paid: %s\n", squad.FromKobo(txn.Amount))
}
```

### Refund Transaction

```go
refund, err := client.Transactions.RefundTransaction(ctx, &squad.RefundTransactionParams{
    GatewayTransactionRef: "gw_ref",
    TransactionRef:        "order-ref-001",
    RefundType:            "Partial", // "Full" or "Partial"
    ReasonForRefund:       "Customer request",
    Amount:                squad.NGN(500),
})
```

### Recurring Payments (Tokenisation)

```go
// Squad includes a ChargeToken in the webhook body after a successful payment.
// Use it to charge the card again without the customer re-entering card details:
resp, err := client.Transactions.InitiatePayment(ctx, &squad.InitiatePaymentParams{
    Email:       "customer@example.com",
    Amount:      squad.NGN(1000),
    Currency:    "NGN",
    IsRecurring: true,
    ChargeToken: &squad.ChargeToken{
        Token:       "tok_abc123",
        ExpiryMonth: 12,
        ExpiryYear:  2027,
    },
})
```

### Iterate Over Missed Webhooks

```go
iter := client.Transactions.AllMissedWebhooks(ctx, nil)
for iter.Next() {
    tx := iter.Item()
    fmt.Println(tx.TransactionRef, tx.Status)
}
if err := iter.Err(); err != nil { log.Fatal(err) }
```

---

## Virtual Accounts

```go
// Create a NUBAN virtual account
account, err := client.VirtualAccounts.Create(ctx, &squad.CreateVirtualAccountParams{
    CustomerIdentifier: "cust-001",
    FirstName:          "Adaeze",
    LastName:           "Okafor",
    MobileNum:          "2348012345678",
    Email:              "adaeze@example.com",
    BVN:                "12345678901",
    DOB:                "01/01/1990",
})
fmt.Println(account.VirtualAccountNumber)

// Iterate over ALL transactions (pages fetched automatically)
iter := client.VirtualAccounts.AllTransactions(ctx, "cust-001", nil)
for iter.Next() {
    tx := iter.Item()
    fmt.Printf("%s credited ₦%.2f from %s\n",
        tx.TransactionRef, squad.FromKobo(tx.Amount), tx.SenderName)
}

// Simulate a credit (sandbox only)
_, err = client.VirtualAccounts.Simulate(ctx, &squad.SimulateVirtualAccountParams{
    VirtualAccountNumber: account.VirtualAccountNumber,
    Amount:               5000,
})
```

---

## Transfers

```go
// Always verify the account before transferring
lookup, err := client.Transfers.AccountLookup(ctx, "057", "0123456789")
fmt.Println("Sending to:", lookup.AccountName)

// Transfer funds
transfer, err := client.Transfers.FundsTransfer(ctx, &squad.FundsTransferParams{
    TransactionRef: "pay-out-001",
    Amount:         squad.NGN(2000),
    BankCode:       "057",
    AccountNumber:  "0123456789",
    AccountName:    lookup.AccountName,
    Currency:       "NGN",
    Remark:         "Salary payment",
})

// Squad-to-Squad transfer
_, err = client.Transfers.IntraTransfer(ctx, &squad.IntraTransferParams{
    TransactionRef:     "intra-001",
    Amount:             squad.NGN(1000),
    SenderIdentifier:   "merchant-A",
    ReceiverIdentifier: "merchant-B",
})

// Iterate over all transfer history
iter := client.Transfers.All(ctx, &squad.TransferListParams{Status: "Success"})
for iter.Next() {
    fmt.Println(iter.Item().TransactionRef, iter.Item().Status)
}
```

---

## Sub-Merchant Management

For aggregators and marketplace platforms that manage vendors or sub-accounts.

```go
// Onboard a new sub-merchant
merchant, err := client.SubMerchants.Create(ctx, &squad.CreateSubMerchantParams{
    DisplayName:   "Vendor Store",
    AccountName:   "Emeka Obi",
    AccountNumber: "0123456789",
    BankCode:      "057",
    Email:         "emeka@vendor.ng",
})
fmt.Println("Sub-merchant ID:", merchant.ID)

// Route a payment through a specific sub-merchant
resp, err := client.Transactions.InitiatePayment(ctx, &squad.InitiatePaymentParams{
    Email:               "buyer@example.com",
    Amount:              squad.NGN(5000),
    Currency:            "NGN",
    InitiatorCustomerID: merchant.MerchantID,
    CallbackURL:         "https://yourplatform.com/callback",
})

// Iterate over all sub-merchants
iter := client.SubMerchants.All(ctx, nil)
for iter.Next() {
    m := iter.Item()
    fmt.Printf("%s — %s\n", m.ID, m.DisplayName)
}

// Remove a sub-merchant
_, err = client.SubMerchants.Delete(ctx, merchant.ID)
```

---

## Disputes

```go
// Iterate over all open disputes
iter := client.Disputes.All(ctx, &squad.DisputeListParams{Status: "open"})
for iter.Next() {
    d := iter.Item()
    fmt.Printf("Ticket: %s | ₦%.2f | %s\n", d.TicketID, squad.FromKobo(d.Amount), d.Reason)
}

// Upload evidence (PDF, PNG, or JPG)
fileData, _ := os.ReadFile("proof-of-delivery.pdf")
_, err = client.Disputes.UploadEvidence(ctx, "ticket-001", fileData, "proof-of-delivery.pdf")

// Reject the dispute (evidence must be uploaded first)
_, err = client.Disputes.RejectDispute(ctx, "ticket-001")

// Or accept it (concede)
_, err = client.Disputes.AcceptDispute(ctx, "ticket-001")
```

---

## Value-Added Services (VAS)

```go
// Buy airtime
_, err := client.VAS.BuyAirtime(ctx, &squad.BuyAirtimeParams{
    PhoneNumber:    "2348012345678",
    Amount:         squad.NGN(50), // minimum ₦50
    Network:        "MTN",         // "MTN", "AIRTEL", "GLO", "9MOBILE"
    TransactionRef: "air-001",
})

// Buy data bundle
plans, _ := client.VAS.GetDataPlans(ctx, "MTN")
_, err = client.VAS.BuyData(ctx, &squad.BuyDataParams{
    PhoneNumber:    "2348012345678",
    PlanCode:       plans.Plans[0].PlanCode,
    Network:        "MTN",
    TransactionRef: "data-001",
})

// Subscribe to cable TV
packages, _ := client.VAS.GetCablePackages(ctx, "DSTV")
_, err = client.VAS.BuyCable(ctx, &squad.BuyCableParams{
    SmartCardNumber: "1234567890",
    PackageCode:     packages.Packages[0].PackageCode,
    Provider:        "DSTV",
    TransactionRef:  "cable-001",
})

// Buy electricity — token returned in response
billers, _ := client.VAS.GetElectricityBillers(ctx)
elec, err := client.VAS.BuyElectricity(ctx, &squad.BuyElectricityParams{
    MeterNumber:    "04123456789",
    Amount:         squad.NGN(5000),
    BillerCode:     billers.Billers[0].BillerCode,
    MeterType:      "prepaid",
    TransactionRef: "elec-001",
})
fmt.Println("Meter token:", elec.ElectricityToken)

// Send SMS
_, err = client.VAS.SendSMS(ctx, &squad.SendSMSParams{
    To:             []string{"2348012345678"},
    From:           "MyBrand",
    Body:           "Your order has been confirmed.",
    TransactionRef: "sms-001",
})
```

---

## Webhooks

### Webhook Router (recommended)

The `WebhookRouter` validates signatures and dispatches events to typed handlers. Register it directly as an `http.Handler`:

```go
router := squad.NewWebhookRouter(os.Getenv("SQUAD_SECRET_KEY")).
    OnTransactionSuccess(func(ctx context.Context, body *squad.WebhookTransactionBody) error {
        return fulfillOrder(body.TransactionRef, body.Amount)
    }).
    OnVirtualAccountCredit(func(ctx context.Context, body *squad.WebhookVirtualAccountBody) error {
        return creditCustomer(body.CustomerIdentifier, body.Amount)
    }).
    OnTransferSuccess(func(ctx context.Context, body *squad.WebhookTransferBody) error {
        return markPayoutComplete(body.TransactionRef)
    }).
    OnDisputeOpened(func(ctx context.Context, body *squad.WebhookDisputeBody) error {
        return notifyTeam(body.TicketID, body.Reason)
    }).
    OnError(func(w http.ResponseWriter, r *http.Request, err error) {
        log.Printf("webhook error: %v", err)
        http.Error(w, "error", http.StatusInternalServerError)
    })

http.Handle("/webhook/squad", router)
```

### Manual Webhook Handling

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    sig := r.Header.Get("x-squad-signature")

    event, err := squad.ParseWebhook(body, sig, os.Getenv("SQUAD_SECRET_KEY"))
    if errors.Is(err, squad.ErrInvalidSignature) {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    switch event.Event {
    case squad.EventTransactionSuccess:
        parsed, _ := event.ParseBody()
        txn := parsed.(*squad.WebhookTransactionBody)
        fmt.Println(txn.TransactionRef, squad.FromKobo(txn.Amount))
    }
    w.WriteHeader(http.StatusOK)
}
```

### Supported Event Types

| Constant | Event String |
|---|---|
| `EventTransactionSuccess` | `charge.success` |
| `EventTransactionFailed` | `charge.failed` |
| `EventVirtualAccountCredit` | `virtual-account.credit` |
| `EventTransferSuccess` | `transfer.success` |
| `EventTransferFailed` | `transfer.failed` |
| `EventTransferReversed` | `transfer.reversed` |
| `EventDisputeOpened` | `dispute.opened` |
| `EventDisputeResolved` | `dispute.resolved` |

---

## Auto-Pagination Iterator

All listing endpoints expose an `All*` / `All` iterator that fetches pages transparently. No more manual pagination loops:

```go
// Without iterator — repetitive, error-prone
page := 1
for {
    result, _ := client.Transfers.GetAllTransactions(ctx, &squad.TransferListParams{Page: page, PerPage: 50})
    for _, t := range result.Transfers { process(t) }
    if len(result.Transfers) < 50 { break }
    page++
}

// With iterator — clean, concise, handles edge cases automatically
iter := client.Transfers.All(ctx, &squad.TransferListParams{PerPage: 50})
for iter.Next() {
    process(iter.Item())
}
if err := iter.Err(); err != nil { log.Fatal(err) }
```

Available iterators:

| Service | Method |
|---|---|
| Transactions | `AllMissedWebhooks(ctx, params)` |
| VirtualAccounts | `AllTransactions(ctx, customerID, params)` |
| Transfers | `All(ctx, params)` |
| Disputes | `All(ctx, params)` |
| SubMerchants | `All(ctx, params)` |

---

## Testing Your Integration with `squadtest`

The `squadtest` package provides a mock Squad API server so you can test your integration without real API calls or a sandbox account.

```bash
go get github.com/kingztech2019/go-squad/squadtest
```

```go
import (
    "testing"
    squad "github.com/kingztech2019/go-squad"
    "github.com/kingztech2019/go-squad/squadtest"
)

func TestMyCheckoutService(t *testing.T) {
    // Start a mock server — automatically shut down when the test ends.
    srv := squadtest.NewServer(t)

    // Register typed handlers that control the response.
    srv.OnInitiatePayment(func(p *squad.InitiatePaymentParams) (*squad.InitiatePaymentResponse, error) {
        // Assert on the request your code sent.
        if p.Amount != squad.NGN(5000) {
            t.Errorf("unexpected amount: %d", p.Amount)
        }
        return &squad.InitiatePaymentResponse{
            CheckoutURL:    "https://fake-checkout.squadco.com/abc",
            TransactionRef: p.TransactionRef,
        }, nil
    })

    srv.OnVerifyTransaction(func(ref string) (*squad.VerifyTransactionResponse, error) {
        return &squad.VerifyTransactionResponse{
            TransactionRef: ref,
            Status:         "Success",
            Amount:         squad.NGN(5000),
        }, nil
    })

    // Inject srv.Client() into the code under test.
    myService := checkout.NewService(srv.Client())

    url, err := myService.StartCheckout("customer@example.com", squad.NGN(5000))
    if err != nil { t.Fatal(err) }
    if url == "" { t.Error("expected checkout URL") }

    // Assert on what your code actually sent.
    if srv.RequestCount() != 1 {
        t.Errorf("expected 1 request, got %d", srv.RequestCount())
    }
}
```

### `squadtest` API

| Method | Description |
|---|---|
| `NewServer(t)` | Create and start a mock server |
| `srv.Client()` | Get a pre-configured `*squad.Client` pointing at the mock |
| `srv.OnInitiatePayment(fn)` | Handle POST /transaction/initiate |
| `srv.OnVerifyTransaction(fn)` | Handle GET /transaction/verify/{ref} |
| `srv.OnRefundTransaction(fn)` | Handle POST /transaction/refund |
| `srv.OnCreateVirtualAccount(fn)` | Handle POST /virtual-account |
| `srv.OnFundsTransfer(fn)` | Handle POST /payout/transfer |
| `srv.OnAccountLookup(fn)` | Handle GET /payout/account/lookup |
| `srv.OnBuyAirtime(fn)` | Handle POST /vas/airtime |
| `srv.OnBuyElectricity(fn)` | Handle POST /vas/electricity |
| `srv.OnCreateSubMerchant(fn)` | Handle POST /merchant/sub-merchant |
| `srv.OnUploadEvidence(fn)` | Handle POST /dispute/upload-evidence/{id} |
| `srv.Handle(method, path, fn)` | Register a custom handler for any endpoint |
| `srv.Requests()` | All recorded requests |
| `srv.LastRequest()` | Most recent request |
| `srv.RequestCount()` | Total requests received |
| `srv.Reset()` | Clear all handlers and recorded requests |

---

## Error Handling

```go
txn, err := client.Transactions.VerifyTransaction(ctx, ref)
if err != nil {
    if squad.IsUnauthorized(err) {
        log.Fatal("invalid API key")
    }
    if squad.IsBadRequest(err) {
        log.Printf("validation error: %v", err)
    }
    if squad.IsNotFound(err) {
        log.Printf("transaction not found: %s", ref)
    }

    // Inspect the full Squad error envelope
    var squadErr *squad.Error
    if errors.As(err, &squadErr) {
        log.Printf("squad status=%d http=%d msg=%s",
            squadErr.Status, squadErr.HTTPStatus, squadErr.Message)
    }
}
```

### Retry Logic

The SDK intentionally does not implement retries — inject a retrying transport instead:

```go
import "github.com/hashicorp/go-retryablehttp"

retryClient := retryablehttp.NewClient()
retryClient.RetryMax = 3

// Always use a stored idempotency key when retrying payments
client := squad.New(secretKey, squad.WithHTTPClient(retryClient.StandardClient()))
```

---

## Testing

```bash
make test          # full test suite with race detector
make test-short    # fast run without race detector
make cover-html    # HTML coverage report
```

---

## Amount Convention

All monetary amounts use the **lowest currency denomination**:

| Currency | Unit | Example |
|---|---|---|
| NGN | kobo (1 NGN = 100 kobo) | `squad.NGN(1000)` → `100000` |
| USD | cents (1 USD = 100 cents) | `squad.USD(10)` → `1000` |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

MIT — see [LICENSE](LICENSE).

---

## Resources

- [Squad API Documentation](https://docs.squadco.com)
- [Squad Dashboard](https://dashboard.squadco.com)
- [Support](mailto:help@squadco.com)
