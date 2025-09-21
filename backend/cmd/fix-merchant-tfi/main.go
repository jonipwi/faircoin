package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/models"
	"faircoin/internal/services"
	"log"

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

	// Initialize fairness service
	fairnessService := services.NewFairnessService(db)

	log.Println("Fixing merchant TFI values...")

	// Get all merchants with TFI = 0
	var merchants []models.User
	if err := db.Where("is_merchant = ? AND tfi = ?", true, 0).Find(&merchants).Error; err != nil {
		log.Fatalf("Failed to get merchants: %v", err)
	}

	log.Printf("Found %d merchants with TFI = 0", len(merchants))

	for _, merchant := range merchants {
		log.Printf("Fixing TFI for merchant %s (ID: %s)", merchant.Username, merchant.ID)

		// Check if merchant has any ratings
		var ratingCount int64
		db.Model(&models.Rating{}).Where("merchant_id = ?", merchant.ID).Count(&ratingCount)

		if ratingCount == 0 {
			// No ratings, set base TFI of 30
			if err := db.Model(&models.User{}).Where("id = ?", merchant.ID).Update("tfi", 30).Error; err != nil {
				log.Printf("Error setting base TFI for merchant %s: %v", merchant.Username, err)
			} else {
				log.Printf("Set base TFI of 30 for merchant %s (no ratings)", merchant.Username)
			}
		} else {
			// Has ratings, recalculate TFI
			if err := fairnessService.UpdateMerchantTFI(merchant.ID); err != nil {
				log.Printf("Error updating TFI for merchant %s: %v", merchant.Username, err)
			} else {
				// Get updated TFI
				var updatedMerchant models.User
				if err := db.First(&updatedMerchant, "id = ?", merchant.ID).Error; err == nil {
					log.Printf("Recalculated TFI for merchant %s: %d (based on %d ratings)",
						merchant.Username, updatedMerchant.TFI, ratingCount)
				}
			}
		}
	}

	// Also fix any merchants who aren't properly flagged as merchants but have ratings
	var usersWithRatings []struct {
		UserID string
		Count  int64
	}

	if err := db.Raw(`
		SELECT users.id as user_id, COUNT(ratings.id) as count 
		FROM users 
		JOIN ratings ON users.id = ratings.merchant_id 
		WHERE users.is_merchant = false 
		GROUP BY users.id
	`).Scan(&usersWithRatings).Error; err != nil {
		log.Printf("Warning: Failed to check for users with ratings who aren't merchants: %v", err)
	} else if len(usersWithRatings) > 0 {
		log.Printf("Found %d users with ratings who aren't flagged as merchants", len(usersWithRatings))
		for _, userWithRating := range usersWithRatings {
			log.Printf("User %s has %d ratings but is not flagged as merchant", userWithRating.UserID, userWithRating.Count)
			// Optionally fix this by setting is_merchant = true and calculating TFI
		}
	}

	log.Println("Merchant TFI fix completed!")

	// Display final merchant statistics
	log.Println("\nFinal merchant statistics:")
	var finalMerchants []models.User
	db.Where("is_merchant = ?", true).Order("tfi DESC").Find(&finalMerchants)

	for _, merchant := range finalMerchants {
		var ratingCount int64
		db.Model(&models.Rating{}).Where("merchant_id = ?", merchant.ID).Count(&ratingCount)
		log.Printf("Merchant %s: TFI = %d, Ratings = %d", merchant.Username, merchant.TFI, ratingCount)
	}
}
