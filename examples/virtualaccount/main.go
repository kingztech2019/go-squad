// Package main demonstrates Squad virtual account creation and management.
package main

import (
	"context"
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

	// Step 1: Create a virtual account for a customer
	fmt.Println("Creating virtual account...")
	account, err := client.VirtualAccounts.Create(context.Background(), &squad.CreateVirtualAccountParams{
		CustomerIdentifier: "cust-456",
		FirstName:          "Adaeze",
		LastName:           "Okafor",
		MobileNum:          "2348012345678",
		Email:              "adaeze@example.com",
		BVN:                "12345678901",
		DOB:                "01/01/1990",
		Gender:             "2",
	})
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Printf("Virtual Account Number: %s\n", account.VirtualAccountNumber)
	fmt.Printf("Customer Identifier: %s\n", account.CustomerIdentifier)

	// Step 2: Query the virtual account details
	fmt.Println("\nQuerying virtual account...")
	details, err := client.VirtualAccounts.Query(context.Background(), account.VirtualAccountNumber)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("Account: %s %s\n", details.FirstName, details.LastName)
	fmt.Printf("Email: %s\n", details.Email)

	// Step 3: Get transaction history
	fmt.Println("\nFetching transactions...")
	txns, err := client.VirtualAccounts.GetTransactions(context.Background(), "cust-456", &squad.VirtualAccountTxParams{
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		log.Fatalf("GetTransactions failed: %v", err)
	}
	fmt.Printf("Total transactions: %d\n", txns.Total)
	for _, tx := range txns.Transactions {
		fmt.Printf("  - Ref: %s | Amount: %d | Sender: %s\n", tx.TransactionRef, tx.Amount, tx.SenderName)
	}

	// Step 4: Sandbox simulation (only works in sandbox mode)
	if account.VirtualAccountNumber != "" {
		fmt.Println("\nSimulating a credit (sandbox only)...")
		sim, err := client.VirtualAccounts.Simulate(context.Background(), &squad.SimulateVirtualAccountParams{
			VirtualAccountNumber: account.VirtualAccountNumber,
			Amount:               5000,
		})
		if err != nil {
			log.Printf("Simulate error (expected in production): %v", err)
		} else {
			fmt.Printf("Simulation Status: %s\n", sim.Status)
		}
	}
}
