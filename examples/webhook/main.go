// Package main demonstrates Squad webhook signature validation and event routing.
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	squad "github.com/kingztech2019/go-squad"
)

func main() {
	secret := os.Getenv("SQUAD_SECRET_KEY")
	if secret == "" {
		log.Fatal("SQUAD_SECRET_KEY environment variable is required")
	}

	http.HandleFunc("/webhook/squad", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		sig := r.Header.Get("x-squad-signature")
		event, err := squad.ParseWebhook(body, sig, secret)
		if errors.Is(err, squad.ErrInvalidSignature) {
			log.Printf("SECURITY: invalid webhook signature from %s", r.RemoteAddr)
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err != nil {
			log.Printf("webhook parse error: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		switch event.Event {
		case squad.EventTransactionSuccess:
			parsed, err := event.ParseBody()
			if err != nil {
				log.Printf("parse body error: %v", err)
				break
			}
			txn := parsed.(*squad.WebhookTransactionBody)
			log.Printf("Payment SUCCESS: ref=%s amount=%d %s customer=%s",
				txn.TransactionRef, txn.Amount, txn.Currency, txn.CustomerEmail)

		case squad.EventTransactionFailed:
			parsed, _ := event.ParseBody()
			txn := parsed.(*squad.WebhookTransactionBody)
			log.Printf("Payment FAILED: ref=%s customer=%s", txn.TransactionRef, txn.CustomerEmail)

		case squad.EventVirtualAccountCredit:
			parsed, _ := event.ParseBody()
			va := parsed.(*squad.WebhookVirtualAccountBody)
			log.Printf("Virtual account CREDITED: account=%s amount=%d sender=%s",
				va.VirtualAccountNumber, va.Amount, va.SenderName)

		case squad.EventTransferSuccess:
			parsed, _ := event.ParseBody()
			t := parsed.(*squad.WebhookTransferBody)
			log.Printf("Transfer SUCCESS: ref=%s amount=%d to=%s (%s)",
				t.TransactionRef, t.Amount, t.AccountName, t.AccountNumber)

		case squad.EventDisputeOpened:
			parsed, _ := event.ParseBody()
			d := parsed.(*squad.WebhookDisputeBody)
			log.Printf("Dispute OPENED: ticket=%s amount=%d reason=%s",
				d.TicketID, d.Amount, d.Reason)

		default:
			log.Printf("Unhandled event type: %s", event.Event)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	})

	addr := ":8080"
	log.Printf("Webhook server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
