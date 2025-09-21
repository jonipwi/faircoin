package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/models"
	"faircoin/internal/services"
	"log"
	"math/rand"

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
	fairnessService := services.NewFairnessService(db)

	log.Println("Setting up merchants and ratings...")

	// Get existing users
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to get users: %v", err)
	}

	if len(users) < 2 {
		log.Fatal("Need at least 2 users to create merchants and ratings")
	}

	log.Printf("Found %d existing users", len(users))

	// Set Carol and Emma as merchants (or first 2 users if they don't exist)
	var merchants []models.User
	var customers []models.User

	for i, user := range users {
		if i < 2 { // First 2 users become merchants
			// Update user to be merchant
			user.IsMerchant = true
			if err := db.Save(&user).Error; err != nil {
				log.Printf("Error setting user %s as merchant: %v", user.Username, err)
				continue
			}
			merchants = append(merchants, user)
			log.Printf("Set %s as merchant", user.Username)
		} else {
			customers = append(customers, user)
		}
	}

	// If we don't have enough customers, use all users as potential customers
	if len(customers) == 0 {
		customers = users
	}

	log.Printf("Merchants: %d, Customers: %d", len(merchants), len(customers))

	// Create ratings for each merchant
	for _, merchant := range merchants {
		log.Printf("Creating ratings for merchant %s...", merchant.Username)

		// Create 5-8 ratings per merchant
		numRatings := 5 + rand.Intn(4)
		ratingsCreated := 0

		for i := 0; i < numRatings && ratingsCreated < numRatings; i++ {
			// Pick a random customer
			customer := customers[rand.Intn(len(customers))]

			// Skip if customer is the same as merchant
			if customer.ID == merchant.ID {
				continue
			}

			// Generate realistic ratings (mostly good with some variation)
			deliveryRating := 7 + rand.Intn(4)      // 7-10
			qualityRating := 6 + rand.Intn(5)       // 6-10
			transparencyRating := 6 + rand.Intn(5)  // 6-10
			environmentalRating := 5 + rand.Intn(6) // 5-10

			comments := []string{
				"Great service, very satisfied!",
				"Good quality products, fast delivery",
				"Excellent communication throughout",
				"Professional and reliable merchant",
				"Will definitely order again",
				"Fast shipping and good packaging",
				"High quality items as described",
				"Responsive to questions",
			}

			_, err := fairnessService.CreateRating(
				customer.ID,
				merchant.ID,
				nil, // No specific transaction ID
				deliveryRating,
				qualityRating,
				transparencyRating,
				environmentalRating,
				comments[rand.Intn(len(comments))],
			)

			if err != nil {
				log.Printf("Error creating rating: %v", err)
			} else {
				log.Printf("Created rating from %s for %s: D:%d Q:%d T:%d E:%d",
					customer.Username, merchant.Username,
					deliveryRating, qualityRating, transparencyRating, environmentalRating)
				ratingsCreated++
			}
		}

		// Update TFI for this merchant
		log.Printf("Updating TFI for merchant %s...", merchant.Username)
		if err := fairnessService.UpdateMerchantTFI(merchant.ID); err != nil {
			log.Printf("Error updating TFI for merchant %s: %v", merchant.Username, err)
		} else {
			// Get updated merchant data
			var updatedMerchant models.User
			if err := db.First(&updatedMerchant, "id = ?", merchant.ID).Error; err == nil {
				log.Printf("Updated TFI for merchant %s: %d", merchant.Username, updatedMerchant.TFI)
			}
		}
	}

	log.Println("Merchant setup completed!")

	// Display final merchant stats
	log.Println("\nFinal merchant statistics:")
	var finalMerchants []models.User
	db.Where("is_merchant = ?", true).Find(&finalMerchants)

	for _, merchant := range finalMerchants {
		var ratingCount int64
		db.Model(&models.Rating{}).Where("merchant_id = ?", merchant.ID).Count(&ratingCount)
		log.Printf("Merchant %s: TFI = %d, Ratings = %d", merchant.Username, merchant.TFI, ratingCount)
	}
}
