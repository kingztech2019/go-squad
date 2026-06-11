// Package main demonstrates Squad payment initiation and verification.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	squad "github.com/kingztech2019/go-squad"
)

func main() {
	secretKey := os.Getenv("SQUAD_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SQUAD_SECRET_KEY environment variable is required")
	}

	client := squad.New(secretKey)

	// Step 1: Initiate a payment
	fmt.Println("Initiating payment...")
	resp, err := client.Transactions.InitiatePayment(context.Background(), &squad.InitiatePaymentParams{
		Email:           "customer@example.com",
		Amount:          500000, // ₦5,000 in kobo
		Currency:        "NGN",
		TransactionRef:  "demo-txn-001",
		CallbackURL:     "https://yoursite.com/payment/callback",
		CustomerName:    "John Doe",
		PaymentChannels: []string{"card", "bank", "ussd", "transfer"},
	})
	if err != nil {
		log.Fatalf("InitiatePayment failed: %v", err)
	}
	fmt.Printf("Checkout URL: %s\n", resp.CheckoutURL)
	fmt.Printf("Transaction Ref: %s\n", resp.TransactionRef)
	fmt.Printf("Total Amount: %.2f NGN\n", resp.TotalAmount/100)

	// Step 2: After the user completes payment, verify the transaction
	fmt.Println("\nVerifying transaction...")
	txn, err := client.Transactions.VerifyTransaction(context.Background(), resp.TransactionRef)
	if err != nil {
		if errors.Is(err, squad.ErrUnauthorized) {
			log.Fatal("Invalid API key")
		}
		log.Fatalf("VerifyTransaction failed: %v", err)
	}
	fmt.Printf("Status: %s\n", txn.Status)
	fmt.Printf("Amount: %d kobo\n", txn.Amount)
	fmt.Printf("Customer: %s (%s)\n", txn.Customer.CustomerName, txn.Customer.CustomerEmail)

	// Step 3: Issue a partial refund (optional)
	if txn.Status == "Success" {
		fmt.Println("\nIssuing partial refund...")
		refund, err := client.Transactions.RefundTransaction(context.Background(), &squad.RefundTransactionParams{
			GatewayTransactionRef: txn.TransactionRef,
			TransactionRef:        txn.TransactionRef,
			RefundType:            "Partial",
			ReasonForRefund:       "Customer requested partial refund",
			Amount:                100000, // ₦1,000 refund
		})
		if err != nil {
			log.Printf("Refund failed: %v", err)
		} else {
			fmt.Printf("Refund Status: %s\n", refund.RefundStatus)
			fmt.Printf("Amount Refunded: %d kobo\n", refund.AmountRefunded)
		}
	}
}
