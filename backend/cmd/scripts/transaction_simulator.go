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
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables - go up two levels from scripts to backend root
	if err := godotenv.Load("../../.env"); err != nil {
		// Try alternate path
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, using environment variables")
		}
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
		log.Fatal("Need at least 2 users to create transactions")
	}

	fmt.Printf("Found %d users, creating 1000 transactions over the past 30 days...\n", len(users))

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Define transaction type probabilities
	transactionTypes := []struct {
		Type           models.TransactionType
		Probability    float64
		MinAmount      float64
		MaxAmount      float64
		RequiresToUser bool
	}{
		{models.TransactionTypeTransfer, 0.60, 5.0, 500.0, true},           // 60% transfers
		{models.TransactionTypeFairnessReward, 0.15, 10.0, 100.0, false},   // 15% fairness rewards
		{models.TransactionTypeMerchantIncentive, 0.10, 5.0, 200.0, false}, // 10% merchant incentives
		{models.TransactionTypeMonthlyIssuance, 0.08, 50.0, 300.0, false},  // 8% monthly issuance
		{models.TransactionTypeFee, 0.05, 0.1, 5.0, false},                 // 5% fees
		{models.TransactionTypeBurn, 0.02, 1.0, 50.0, false},               // 2% burns
	}

	// Create 1000 transactions
	for i := 0; i < 1000; i++ {
		// Random date within the past 30 days
		daysAgo := rand.Intn(30)
		hoursAgo := rand.Intn(24)
		minutesAgo := rand.Intn(60)
		secondsAgo := rand.Intn(60)

		transactionDate := time.Now().
			AddDate(0, 0, -daysAgo).
			Add(-time.Duration(hoursAgo) * time.Hour).
			Add(-time.Duration(minutesAgo) * time.Minute).
			Add(-time.Duration(secondsAgo) * time.Second)

		// Select transaction type based on probability
		txType := selectTransactionType(transactionTypes)

		// Select random user as sender
		fromUserIdx := rand.Intn(len(users))
		fromUser := users[fromUserIdx]

		// Select receiver if needed
		var toUser *models.User
		if txType.RequiresToUser {
			toUserIdx := rand.Intn(len(users))
			for toUserIdx == fromUserIdx {
				toUserIdx = rand.Intn(len(users))
			}
			toUser = &users[toUserIdx]
		}

		// Generate random amount within type's range
		amount := txType.MinAmount + rand.Float64()*(txType.MaxAmount-txType.MinAmount)
		amount = float64(int(amount*100)) / 100 // Round to 2 decimal places

		// Calculate fee (0.1% for most transactions)
		fee := 0.0
		if txType.Type == models.TransactionTypeTransfer || txType.Type == models.TransactionTypeMerchantIncentive {
			fee = amount * 0.001
		}

		// Generate description based on transaction type
		description := generateDescription(txType.Type, fromUser.Username, toUser)

		// Create transaction
		transaction := &models.Transaction{
			ID:          uuid.New(),
			UserID:      fromUser.ID,
			Type:        txType.Type,
			Amount:      amount,
			Fee:         fee,
			Description: description,
			Status:      selectStatus(),
			CreatedAt:   transactionDate,
		}

		if toUser != nil {
			transaction.ToUserID = &toUser.ID
		}

		if err := db.Create(transaction).Error; err != nil {
			log.Printf("Error creating transaction %d: %v", i+1, err)
		} else {
			if i%100 == 0 {
				fmt.Printf("Created %d transactions...\n", i+1)
			}
		}
	}

	fmt.Println("\nTransaction simulation completed!")

	// Show summary statistics
	showSummaryStats(db)
}

func selectTransactionType(types []struct {
	Type           models.TransactionType
	Probability    float64
	MinAmount      float64
	MaxAmount      float64
	RequiresToUser bool
}) struct {
	Type           models.TransactionType
	Probability    float64
	MinAmount      float64
	MaxAmount      float64
	RequiresToUser bool
} {
	r := rand.Float64()
	cumulative := 0.0

	for _, txType := range types {
		cumulative += txType.Probability
		if r <= cumulative {
			return txType
		}
	}

	// Fallback to transfer
	return types[0]
}

func generateDescription(txType models.TransactionType, fromUser string, toUser *models.User) string {
	switch txType {
	case models.TransactionTypeTransfer:
		if toUser != nil {
			return fmt.Sprintf("Transfer from %s to %s", fromUser, toUser.Username)
		}
		return "Transfer transaction"
	case models.TransactionTypeFairnessReward:
		rewards := []string{
			"Weekly fairness reward",
			"Community service reward",
			"Peer rating bonus",
			"Environmental action reward",
			"Dispute resolution reward",
		}
		return rewards[rand.Intn(len(rewards))]
	case models.TransactionTypeMerchantIncentive:
		incentives := []string{
			"New merchant signup bonus",
			"High-volume merchant reward",
			"Quality service incentive",
			"Sustainable business reward",
		}
		return incentives[rand.Intn(len(incentives))]
	case models.TransactionTypeMonthlyIssuance:
		return "Monthly coin issuance"
	case models.TransactionTypeFee:
		return "Transaction processing fee"
	case models.TransactionTypeBurn:
		return "Token burn for deflation"
	default:
		return "Transaction"
	}
}

func selectStatus() string {
	statuses := []string{"completed", "pending", "failed"}
	weights := []float64{0.85, 0.10, 0.05} // 85% completed, 10% pending, 5% failed

	r := rand.Float64()
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return statuses[i]
		}
	}

	return "completed"
}

func showSummaryStats(db *gorm.DB) {
	fmt.Println("\n=== TRANSACTION SUMMARY ===")

	// Total transactions
	var totalCount int64
	db.Model(&models.Transaction{}).Count(&totalCount)
	fmt.Printf("Total transactions in database: %d\n", totalCount)

	// Transactions by type
	fmt.Println("\nTransactions by type:")
	types := []models.TransactionType{
		models.TransactionTypeTransfer,
		models.TransactionTypeFairnessReward,
		models.TransactionTypeMerchantIncentive,
		models.TransactionTypeMonthlyIssuance,
		models.TransactionTypeFee,
		models.TransactionTypeBurn,
	}

	for _, txType := range types {
		var count int64
		var totalAmount float64

		db.Model(&models.Transaction{}).Where("type = ?", txType).Count(&count)
		db.Model(&models.Transaction{}).Where("type = ?", txType).Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalAmount)

		fmt.Printf("  %s: %d transactions, %.2f FC total\n", txType, count, totalAmount)
	}

	// Transactions by status
	fmt.Println("\nTransactions by status:")
	statuses := []string{"completed", "pending", "failed"}
	for _, status := range statuses {
		var count int64
		db.Model(&models.Transaction{}).Where("status = ?", status).Count(&count)
		fmt.Printf("  %s: %d transactions\n", status, count)
	}

	// Recent activity (last 7 days)
	fmt.Println("\nRecent activity (last 7 days):")
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		dayEnd := dayStart.Add(24 * time.Hour)

		var dayStats struct {
			Count  int64
			Volume float64
		}

		db.Model(&models.Transaction{}).
			Where("created_at BETWEEN ? AND ? AND status = 'completed'", dayStart, dayEnd).
			Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as volume").
			Row().Scan(&dayStats.Count, &dayStats.Volume)

		fmt.Printf("  %s: %d transactions, %.2f FC\n",
			date.Format("Jan 02"), dayStats.Count, dayStats.Volume)
	}

	// Updated monthly volume (last 6 months including new data)
	fmt.Println("\nMonthly transfer volume (last 6 months):")
	for i := 5; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		var monthVolume struct {
			Total float64
			Count int64
		}

		db.Model(&models.Transaction{}).
			Where("created_at BETWEEN ? AND ? AND type = ? AND status = 'completed'",
				monthStart, monthEnd, models.TransactionTypeTransfer).
			Select("COALESCE(SUM(amount), 0) as total, COUNT(*) as count").
			Row().Scan(&monthVolume.Total, &monthVolume.Count)

		fmt.Printf("  %s: %.2f FC (%d transactions)\n",
			date.Format("Jan 2006"), monthVolume.Total, monthVolume.Count)
	}
}
