package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/models"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Check total number of transactions
	var totalCount int64
	db.Model(&models.Transaction{}).Count(&totalCount)
	fmt.Printf("Total transactions in database: %d\n", totalCount)

	// Check transfer transactions specifically
	var transferCount int64
	db.Model(&models.Transaction{}).Where("type = ?", models.TransactionTypeTransfer).Count(&transferCount)
	fmt.Printf("Transfer transactions: %d\n", transferCount)

	// Check all transaction types
	var allTransactions []models.Transaction
	db.Find(&allTransactions)
	fmt.Printf("All transactions:\n")
	for _, tx := range allTransactions {
		fmt.Printf("  ID: %s, Type: %s, Amount: %.2f, Date: %s\n", tx.ID, tx.Type, tx.Amount, tx.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Check transactions by month for the last 6 months
	fmt.Println("\nTransaction volume by month (last 6 months):")
	for i := 5; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		var monthVolume struct {
			Total float64
			Count int64
		}

		db.Model(&models.Transaction{}).
			Where("created_at BETWEEN ? AND ? AND type = ?", monthStart, monthEnd, models.TransactionTypeTransfer).
			Select("COALESCE(SUM(amount), 0) as total, COUNT(*) as count").Scan(&monthVolume)

		fmt.Printf("%s: %.2f FC (%d transactions)\n", date.Format("Jan 2006"), monthVolume.Total, monthVolume.Count)
	}

	// Show some sample transactions
	fmt.Println("\nSample transfer transactions:")
	var transactions []models.Transaction
	db.Where("type = ?", models.TransactionTypeTransfer).Limit(5).Find(&transactions)

	for _, tx := range transactions {
		fmt.Printf("ID: %s, Amount: %.2f, Date: %s\n", tx.ID, tx.Amount, tx.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}
