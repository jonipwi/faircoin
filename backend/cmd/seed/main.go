package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/models"
	"faircoin/internal/services"
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

	// Initialize services
	userService := services.NewUserService(db)
	walletService := services.NewWalletService(db)
	transactionService := services.NewTransactionService(db)

	// Create test data
	log.Println("Creating test users...")

	testUsers := []struct {
		username  string
		email     string
		firstName string
		lastName  string
		pfi       int
	}{
		{"alice123", "alice@example.com", "Alice", "Johnson", 85},
		{"bob456", "bob@example.com", "Bob", "Smith", 67},
		{"carol789", "carol@example.com", "Carol", "Davis", 94},
		{"david999", "david@example.com", "David", "Wilson", 72},
		{"emma555", "emma@example.com", "Emma", "Brown", 88},
	}

	var userIDs []uuid.UUID

	for _, testUser := range testUsers {
		user, err := userService.CreateUser(
			testUser.username,
			testUser.email,
			"password123",
			testUser.firstName,
			testUser.lastName,
		)
		if err != nil {
			log.Printf("Error creating user %s: %v", testUser.username, err)
			continue
		}

		// Update PFI score
		err = userService.UpdateUser(user.ID, map[string]interface{}{
			"pfi":         testUser.pfi,
			"is_verified": true,
		})
		if err != nil {
			log.Printf("Error updating user %s: %v", testUser.username, err)
		}

		userIDs = append(userIDs, user.ID)
		log.Printf("Created user: %s", testUser.username)
	}

	// Create some test transactions
	log.Println("Creating test transactions...")

	if len(userIDs) >= 2 {
		for i := 0; i < 10; i++ {
			fromIdx := rand.Intn(len(userIDs))
			toIdx := rand.Intn(len(userIDs))
			if fromIdx == toIdx {
				continue
			}

			amount := float64(rand.Intn(1000) + 10)
			_, err := walletService.Transfer(
				userIDs[fromIdx],
				userIDs[toIdx],
				amount,
				"Test transaction",
			)
			if err != nil {
				log.Printf("Error creating transaction: %v", err)
				continue
			}
			log.Printf("Created transaction: %.2f FC", amount)
		}
	}

	// Create some transfer transactions with different dates for volume chart
	log.Println("Creating transfer transactions over time...")
	if len(userIDs) >= 2 {
		for month := 5; month >= 0; month-- {
			for i := 0; i < 3; i++ { // 3 transactions per month
				fromIdx := rand.Intn(len(userIDs))
				toIdx := rand.Intn(len(userIDs))
				if fromIdx == toIdx {
					continue
				}

				amount := float64(rand.Intn(50) + 20) // Smaller amounts to avoid balance issues

				// Create transaction record directly in database (bypassing balance checks)
				transaction := &models.Transaction{
					UserID:      userIDs[fromIdx],
					ToUserID:    &userIDs[toIdx],
					Type:        models.TransactionTypeTransfer,
					Amount:      amount,
					Fee:         amount * 0.001,
					Description: "Historical test transaction",
					Status:      "completed",
					CreatedAt:   time.Now().AddDate(0, -month, -rand.Intn(28)),
				}

				if err := transactionService.GetDB().Create(transaction).Error; err != nil {
					log.Printf("Error creating transfer transaction: %v", err)
				} else {
					log.Printf("Created transfer transaction: %.2f FC for month -%d", amount, month)
				}
			}
		}
	}

	// Create some fairness rewards
	log.Println("Creating fairness rewards...")
	for i, userID := range userIDs {
		if i >= 3 { // Only for first 3 users
			break
		}

		// Create fairness reward transaction
		transaction := &models.Transaction{
			UserID:      userID,
			Type:        models.TransactionTypeFairnessReward,
			Amount:      float64(rand.Intn(50) + 10),
			Fee:         0,
			Description: "Monthly fairness reward",
			Status:      "completed",
			CreatedAt:   time.Now().AddDate(0, 0, -rand.Intn(7)),
		}

		if err := transactionService.GetDB().Create(transaction).Error; err != nil {
			log.Printf("Error creating fairness reward: %v", err)
		} else {
			log.Printf("Created fairness reward: %.2f FC", transaction.Amount)
		}
	}

	log.Println("Test data creation completed!")
}
