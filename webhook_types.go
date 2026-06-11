package squad

import (
	"encoding/json"
	"fmt"
)

// EventType is the string identifier for a Squad webhook event.
type EventType string

// Squad webhook event type constants.
const (
	EventTransactionSuccess   EventType = "charge.success"
	EventTransactionFailed    EventType = "charge.failed"
	EventVirtualAccountCredit EventType = "virtual-account.credit"
	EventTransferSuccess      EventType = "transfer.success"
	EventTransferFailed       EventType = "transfer.failed"
	EventTransferReversed     EventType = "transfer.reversed"
	EventDisputeOpened        EventType = "dispute.opened"
	EventDisputeResolved      EventType = "dispute.resolved"
)

// WebhookEvent is the top-level structure of all Squad webhook payloads.
// Use event.Event to determine the event type, then call ParseBody() to decode the body.
type WebhookEvent struct {
	Event EventType       `json:"event"`
	Body  json.RawMessage `json:"body"`
}

// ParseBody decodes the Body field into the appropriate typed struct based on Event.
// Returns one of:
//   - *WebhookTransactionBody    for EventTransactionSuccess / EventTransactionFailed
//   - *WebhookVirtualAccountBody for EventVirtualAccountCredit
//   - *WebhookTransferBody       for EventTransferSuccess / EventTransferFailed / EventTransferReversed
//   - *WebhookDisputeBody        for EventDisputeOpened / EventDisputeResolved
//   - json.RawMessage            for unknown or future event types
func (e *WebhookEvent) ParseBody() (any, error) {
	switch e.Event {
	case EventTransactionSuccess, EventTransactionFailed:
		var body WebhookTransactionBody
		if err := json.Unmarshal(e.Body, &body); err != nil {
			return nil, fmt.Errorf("squad: decode transaction webhook body: %w", err)
		}
		return &body, nil

	case EventVirtualAccountCredit:
		var body WebhookVirtualAccountBody
		if err := json.Unmarshal(e.Body, &body); err != nil {
			return nil, fmt.Errorf("squad: decode virtual account webhook body: %w", err)
		}
		return &body, nil

	case EventTransferSuccess, EventTransferFailed, EventTransferReversed:
		var body WebhookTransferBody
		if err := json.Unmarshal(e.Body, &body); err != nil {
			return nil, fmt.Errorf("squad: decode transfer webhook body: %w", err)
		}
		return &body, nil

	case EventDisputeOpened, EventDisputeResolved:
		var body WebhookDisputeBody
		if err := json.Unmarshal(e.Body, &body); err != nil {
			return nil, fmt.Errorf("squad: decode dispute webhook body: %w", err)
		}
		return &body, nil

	default:
		return e.Body, nil
	}
}

// WebhookTransactionBody holds the body payload for charge.success and charge.failed events.
type WebhookTransactionBody struct {
	TransactionRef string         `json:"transaction_ref"`
	Amount         int64          `json:"amount"`
	Currency       string         `json:"currency"`
	Status         string         `json:"transaction_status"`
	Channel        string         `json:"channel"`
	CustomerEmail  string         `json:"customer_email"`
	CustomerName   string         `json:"customer_name"`
	Meta           map[string]any `json:"meta,omitempty"`
	GatewayRef     string         `json:"gateway_ref"`
	IsRecurring    bool           `json:"is_recurring"`
	ChargeToken    *ChargeToken   `json:"charge_token,omitempty"`
	CreatedAt      string         `json:"created_at"`
}

// WebhookVirtualAccountBody holds the body payload for virtual-account.credit events.
type WebhookVirtualAccountBody struct {
	VirtualAccountNumber string `json:"virtual_account_number"`
	Amount               int64  `json:"amount"`
	Currency             string `json:"currency"`
	SenderName           string `json:"sender_name"`
	SenderBank           string `json:"sender_bank"`
	TransactionRef       string `json:"transaction_ref"`
	CustomerIdentifier   string `json:"customer_identifier"`
	CreatedAt            string `json:"created_at"`
}

// WebhookTransferBody holds the body payload for transfer events.
type WebhookTransferBody struct {
	TransactionRef string `json:"transaction_ref"`
	Amount         int64  `json:"amount"`
	Status         string `json:"status"`
	AccountName    string `json:"account_name"`
	AccountNumber  string `json:"account_number"`
	BankCode       string `json:"bank_code"`
	CreatedAt      string `json:"created_at"`
}

// WebhookDisputeBody holds the body payload for dispute events.
type WebhookDisputeBody struct {
	TicketID       string `json:"ticket_id"`
	TransactionRef string `json:"transaction_ref"`
	Amount         int64  `json:"amount"`
	Status         string `json:"dispute_status"`
	Reason         string `json:"dispute_reason"`
	CreatedAt      string `json:"created_at"`
}
