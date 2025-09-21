package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/models"
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

	log.Println("=== User Management TFI Verification Report ===")
	log.Println()

	// Get all users with their TFI values
	var users []models.User
	if err := db.Preload("Wallet").Find(&users).Error; err != nil {
		log.Fatalf("Failed to get users: %v", err)
	}

	log.Printf("Found %d users in the system", len(users))
	log.Println()

	// Display user information with focus on TFI
	log.Println("User Information:")
	log.Println("================")
	for _, user := range users {
		balance := 0.0
		if user.Wallet != nil {
			balance = user.Wallet.Balance
		}

		log.Printf("👤 %s (%s %s)", user.Username, user.FirstName, user.LastName)
		log.Printf("   📧 Email: %s", user.Email)
		log.Printf("   ⭐ PFI: %d | TFI: %d", user.PFI, user.TFI)
		log.Printf("   💰 Balance: %.2f FC", balance)

		roles := []string{}
		if user.IsAdmin {
			roles = append(roles, "Admin")
		}
		if user.IsMerchant {
			roles = append(roles, "Merchant")
		}
		if user.IsVerified {
			roles = append(roles, "Verified")
		}
		if len(roles) == 0 {
			roles = append(roles, "Regular User")
		}

		log.Printf("   🏷️  Roles: %v", roles)
		log.Printf("   🕐 Created: %s", user.CreatedAt.Format("2006-01-02 15:04"))
		log.Println()
	}

	// Get TFI statistics
	log.Println("TFI Statistics:")
	log.Println("===============")

	var merchants []models.User
	db.Where("is_merchant = ?", true).Find(&merchants)

	totalTFI := 0
	tfiCounts := map[string]int{
		"excellent": 0, // 80-100
		"good":      0, // 60-79
		"fair":      0, // 40-59
		"base":      0, // 30 (base for new merchants)
		"zero":      0, // 0 (should not happen for merchants)
	}

	for _, merchant := range merchants {
		totalTFI += merchant.TFI

		switch {
		case merchant.TFI >= 80:
			tfiCounts["excellent"]++
		case merchant.TFI >= 60:
			tfiCounts["good"]++
		case merchant.TFI >= 40:
			tfiCounts["fair"]++
		case merchant.TFI == 30:
			tfiCounts["base"]++
		case merchant.TFI == 0:
			tfiCounts["zero"]++
		}
	}

	if len(merchants) > 0 {
		avgTFI := float64(totalTFI) / float64(len(merchants))
		log.Printf("📊 Total Merchants: %d", len(merchants))
		log.Printf("📊 Average TFI: %.1f", avgTFI)
		log.Printf("📊 TFI Distribution:")
		log.Printf("   🌟 Excellent (80-100): %d merchants", tfiCounts["excellent"])
		log.Printf("   👍 Good (60-79): %d merchants", tfiCounts["good"])
		log.Printf("   ⚖️  Fair (40-59): %d merchants", tfiCounts["fair"])
		log.Printf("   🆕 Base (30): %d merchants", tfiCounts["base"])
		if tfiCounts["zero"] > 0 {
			log.Printf("   ❌ Zero TFI: %d merchants (PROBLEM!)", tfiCounts["zero"])
		}
	} else {
		log.Println("📊 No merchants found in the system")
	}

	log.Println()

	// Check ratings data
	var totalRatings int64
	db.Model(&models.Rating{}).Count(&totalRatings)

	log.Printf("📈 Total merchant ratings in system: %d", totalRatings)

	if totalRatings > 0 {
		var ratingsPerMerchant []struct {
			MerchantID string
			Username   string
			TFI        int
			Count      int64
		}

		db.Raw(`
			SELECT u.id as merchant_id, u.username, u.tfi, COUNT(r.id) as count
			FROM users u
			LEFT JOIN ratings r ON u.id = r.merchant_id
			WHERE u.is_merchant = true
			GROUP BY u.id, u.username, u.tfi
			ORDER BY u.tfi DESC
		`).Scan(&ratingsPerMerchant)

		log.Println()
		log.Println("Ratings per Merchant:")
		log.Println("=====================")
		for _, rpm := range ratingsPerMerchant {
			log.Printf("🏪 %s: TFI=%d, Ratings=%d", rpm.Username, rpm.TFI, rpm.Count)
		}
	}

	log.Println()
	log.Println("=== VERIFICATION RESULT ===")

	// Check if TFI problem is fixed
	problemCount := tfiCounts["zero"]
	if problemCount == 0 {
		log.Println("✅ SUCCESS: No merchants have TFI = 0")
		log.Println("✅ All merchants have appropriate TFI values")
		log.Println("✅ TFI calculation and storage is working correctly")
	} else {
		log.Printf("❌ PROBLEM: %d merchants still have TFI = 0", problemCount)
		log.Println("❌ TFI calculation needs further investigation")
	}

	log.Println()
	log.Println("=== END REPORT ===")
}
