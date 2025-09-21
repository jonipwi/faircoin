package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/models"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
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

	// Get existing users
	var users []models.User
	db.Find(&users)

	if len(users) < 2 {
		log.Fatal("Need at least 2 users to create transfer transactions")
	}

	fmt.Printf("Found %d users, creating transfer transactions...\n", len(users))

	// Create transfer transactions for the past 6 months
	for month := 5; month >= 0; month-- {
		transactionsPerMonth := rand.Intn(5) + 3 // 3-7 transactions per month

		for i := 0; i < transactionsPerMonth; i++ {
			fromIdx := rand.Intn(len(users))
			toIdx := rand.Intn(len(users))
			for fromIdx == toIdx {
				toIdx = rand.Intn(len(users))
			}

			amount := float64(rand.Intn(100) + 20) // Random amount between 20-120
			fee := amount * 0.001

			// Create date within the month
			baseDate := time.Now().AddDate(0, -month, 0)
			dayOffset := rand.Intn(28) // Random day in month
			transactionDate := time.Date(baseDate.Year(), baseDate.Month(), dayOffset+1,
				rand.Intn(24), rand.Intn(60), rand.Intn(60), 0, baseDate.Location())

			transaction := &models.Transaction{
				ID:          uuid.New(),
				UserID:      users[fromIdx].ID,
				ToUserID:    &users[toIdx].ID,
				Type:        models.TransactionTypeTransfer,
				Amount:      amount,
				Fee:         fee,
				Description: "Transfer transaction",
				Status:      "completed",
				CreatedAt:   transactionDate,
			}

			if err := db.Create(transaction).Error; err != nil {
				log.Printf("Error creating transfer transaction: %v", err)
			} else {
				fmt.Printf("Created transfer: %.2f FC from %s to %s on %s\n",
					amount, users[fromIdx].Username, users[toIdx].Username,
					transactionDate.Format("2006-01-02"))
			}
		}
	}

	// Verify the data
	fmt.Println("\nVerifying transaction volume by month:")
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

	fmt.Println("\nTransaction volume data creation completed!")
}
