package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/models"
	"faircoin/internal/services"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
)

// Demo represents the FairCoin demonstration
type Demo struct {
	db                 *gorm.DB
	userService        *services.UserService
	walletService      *services.WalletService
	transactionService *services.TransactionService
	fairnessService    *services.FairnessService
	users              []models.User
}

func main() {
	fmt.Println("🚀 FairCoin PFI/TFI Demonstration System")
	fmt.Println("========================================")

	demo := &Demo{}
	if err := demo.initialize(); err != nil {
		log.Fatalf("Failed to initialize demo: %v", err)
	}

	// Run the complete demonstration
	demo.runDemo()
}

func (d *Demo) initialize() error {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	var err error
	d.db, err = database.Initialize(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	d.userService = services.NewUserService(d.db)
	d.walletService = services.NewWalletService(d.db)
	d.transactionService = services.NewTransactionService(d.db)
	d.fairnessService = services.NewFairnessService(d.db)

	return nil
}

func (d *Demo) runDemo() {
	defer d.db.Close()

	fmt.Println("\n📋 Demo Steps:")
	fmt.Println("1. Clear existing database")
	fmt.Println("2. Create diverse user profiles")
	fmt.Println("3. Generate community activities & attestations")
	fmt.Println("4. Simulate merchant transactions & ratings")
	fmt.Println("5. Calculate PFI★ and TFI★ scores")
	fmt.Println("6. Generate transaction history")
	fmt.Println("7. Create governance proposals")
	fmt.Println("8. Display final results")

	// Step 1: Clear database
	d.clearDatabase()

	// Step 2: Create diverse users
	d.createDiverseUsers()

	// Step 3: Generate community activities
	d.generateCommunityActivities()

	// Step 4: Simulate merchant activities
	d.simulateMerchantActivities()

	// Step 5: Calculate fairness scores
	d.calculateFairnessScores()

	// Step 6: Generate transaction history
	d.generateTransactionHistory()

	// Step 7: Create governance proposals
	d.createGovernanceProposals()

	// Step 8: Display results
	d.displayResults()

	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Println("💡 Access the admin dashboard at: http://localhost:8080")
	fmt.Println("🔐 Admin login: admin / password123")
	fmt.Println("🔐 User login: alice123 / password123")
}

func (d *Demo) clearDatabase() {
	fmt.Println("\n🗑️  Step 1: Clearing existing database...")

	// Clear all data from tables in proper order (respecting foreign keys)
	tables := []string{
		"votes",
		"proposals",
		"ratings",
		"attestations",
		"transactions",
		"wallets",
		"monetary_policies",
		"community_basket_indices",
		"users",
	}

	for _, table := range tables {
		if err := d.db.Exec("DELETE FROM " + table).Error; err != nil {
			log.Printf("Warning: Could not clear table %s: %v", table, err)
		}
	}

	// Reset the users slice
	d.users = nil

	// Ensure tables exist (recreate if needed)
	if err := database.Migrate(d.db); err != nil {
		// For SQLite with existing tables, migration errors are expected
		// Just log a warning and continue
		log.Printf("Migration warning (expected for existing database): %v", err)
		fmt.Println("   ⚠️  Database migration had warnings (tables already exist)")
	}

	fmt.Println("   ✅ Database cleared and recreated")
}

func (d *Demo) createDiverseUsers() {
	fmt.Println("\n👥 Step 2: Creating diverse user profiles...")

	userProfiles := []struct {
		username    string
		email       string
		firstName   string
		lastName    string
		isMerchant  bool
		isAdmin     bool
		description string
		expectedPFI string
		expectedTFI string
	}{
		{"admin", "admin@faircoin.local", "System", "Administrator", false, true, "FairCoin system administrator", "N/A", "N/A"},
		{"alice123", "alice@example.com", "Alice", "Johnson", false, false, "Community volunteer & environmental activist", "High PFI", "N/A"},
		{"bob456", "bob@example.com", "Bob", "Smith", true, false, "Local organic farmer & fair trade merchant", "Medium PFI", "High TFI"},
		{"carol789", "carol@example.com", "Carol", "Davis", false, false, "Social worker & dispute mediator", "Very High PFI", "N/A"},
		{"david999", "david@example.com", "David", "Wilson", true, false, "Tech startup founder with sustainable practices", "Medium PFI", "Medium TFI"},
		{"emma555", "emma@example.com", "Emma", "Brown", false, false, "University student & peer tutor", "High PFI", "N/A"},
		{"frank777", "frank@example.com", "Frank", "Miller", true, false, "Traditional retailer, new to fairness", "Low PFI", "Low TFI"},
		{"grace888", "grace@example.com", "Grace", "Lee", false, false, "Retired teacher, community elder", "Very High PFI", "N/A"},
		{"henry111", "henry@example.com", "Henry", "Garcia", true, false, "Renewable energy consultant", "High PFI", "High TFI"},
	}

	for _, profile := range userProfiles {
		user, err := d.userService.CreateUser(
			profile.username,
			profile.email,
			"password123",
			profile.firstName,
			profile.lastName,
		)
		if err != nil {
			log.Printf("Error creating user %s: %v", profile.username, err)
			continue
		}

		// Update user roles and status
		updateData := map[string]interface{}{
			"is_verified": true,
		}

		if profile.isMerchant {
			updateData["is_merchant"] = true
			user.IsMerchant = true // Update local copy
		}

		if profile.isAdmin {
			updateData["is_admin"] = true
			user.IsAdmin = true // Update local copy
		}

		err = d.userService.UpdateUser(user.ID, updateData)
		if err != nil {
			log.Printf("Error updating user %s: %v", profile.username, err)
		}

		user.IsVerified = true // Update local copy

		d.users = append(d.users, *user)

		roleDescription := ""
		if profile.isAdmin {
			roleDescription = " [ADMIN]"
		} else if profile.isMerchant {
			roleDescription = " [MERCHANT]"
		}

		fmt.Printf("   👤 %s (%s)%s - %s\n", profile.firstName, profile.username, roleDescription, profile.description)
		if !profile.isAdmin {
			fmt.Printf("      Expected: %s", profile.expectedPFI)
			if profile.expectedTFI != "N/A" {
				fmt.Printf(", %s", profile.expectedTFI)
			}
			fmt.Println()
		}
	}

	fmt.Printf("   ✅ Created %d diverse users\n", len(d.users))
}

func (d *Demo) generateCommunityActivities() {
	fmt.Println("\n🤝 Step 3: Generating community activities & peer attestations...")

	// Check if we have users to generate activities for
	if len(d.users) == 0 {
		fmt.Println("   ⚠️  No users found, skipping community activities generation")
		return
	}

	// Community service activities (excluding admin user at index 0)
	communityActivities := []struct {
		userIndex  int
		hours      int
		activities []string
	}{
		{1, 45, []string{"Environmental cleanup", "Tree planting", "Recycling program"}},            // Alice
		{2, 25, []string{"Community garden", "Farmers market organization"}},                        // Bob
		{3, 60, []string{"Dispute mediation", "Community counseling", "Youth mentoring"}},           // Carol
		{4, 15, []string{"Tech workshops", "Digital literacy training"}},                            // David
		{5, 35, []string{"Peer tutoring", "Study groups", "Academic support"}},                      // Emma
		{6, 5, []string{"Neighborhood watch"}},                                                      // Frank
		{7, 80, []string{"Senior center activities", "Wisdom sharing", "Community history"}},        // Grace
		{8, 40, []string{"Solar panel installations", "Energy audits", "Sustainability workshops"}}, // Henry
	}

	for _, activity := range communityActivities {
		if activity.userIndex < len(d.users) {
			user := d.users[activity.userIndex]

			// Update community service hours
			err := d.userService.UpdateUser(user.ID, map[string]interface{}{
				"community_service": activity.hours,
			})
			if err != nil {
				log.Printf("Error updating community service for %s: %v", user.Username, err)
				continue
			}

			fmt.Printf("   🏆 %s: %d hours - %v\n", user.FirstName, activity.hours, activity.activities)
		}
	}

	// Generate peer attestations
	d.generatePeerAttestations()

	fmt.Println("   ✅ Community activities and attestations generated")
}

func (d *Demo) generatePeerAttestations() {
	// Create realistic peer attestations (excluding admin user at index 0)
	attestations := []struct {
		userIndex     int
		attesterIndex int
		type_         string
		value         int
		description   string
	}{
		// Alice attestations (environmental activist)
		{1, 3, "community_service", 9, "Outstanding environmental leadership"},
		{1, 5, "community_service", 8, "Inspiring recycling initiative"},
		{1, 7, "peer_rating", 9, "Always helpful and reliable"},

		// Bob attestations (organic farmer)
		{2, 1, "community_service", 7, "Great community garden work"},
		{2, 8, "peer_rating", 8, "Trustworthy and hardworking"},

		// Carol attestations (social worker)
		{3, 1, "dispute_resolution", 10, "Excellent mediation skills"},
		{3, 2, "dispute_resolution", 9, "Fair and impartial mediator"},
		{3, 5, "community_service", 9, "Dedicated community service"},
		{3, 7, "peer_rating", 10, "Most trustworthy person I know"},

		// Emma attestations (student)
		{5, 1, "community_service", 8, "Excellent peer tutoring"},
		{5, 3, "peer_rating", 8, "Helpful and patient tutor"},

		// Grace attestations (community elder)
		{7, 1, "community_service", 10, "Lifetime of community service"},
		{7, 3, "community_service", 10, "Wisdom and guidance for all"},
		{7, 5, "peer_rating", 9, "Community treasure"},

		// Henry attestations (renewable energy)
		{8, 1, "community_service", 8, "Great sustainability work"},
		{8, 7, "peer_rating", 8, "Knowledgeable and helpful"},
	}

	for _, att := range attestations {
		if att.userIndex < len(d.users) && att.attesterIndex < len(d.users) {
			attestation := &models.Attestation{
				UserID:      d.users[att.userIndex].ID,
				AttesterID:  d.users[att.attesterIndex].ID,
				Type:        att.type_,
				Value:       att.value,
				Description: att.description,
				Verified:    true,
				CreatedAt:   time.Now().AddDate(0, 0, -rand.Intn(30)),
			}

			if err := d.db.Create(attestation).Error; err != nil {
				log.Printf("Error creating attestation: %v", err)
			}
		}
	}
}

func (d *Demo) simulateMerchantActivities() {
	fmt.Println("\n🛍️  Step 4: Simulating merchant transactions & customer ratings...")

	// Check if we have users to simulate merchant activities for
	if len(d.users) == 0 {
		fmt.Println("   ⚠️  No users found, skipping merchant activities generation")
		return
	}

	// Get merchants
	merchants := []models.User{}
	for _, user := range d.users {
		if user.IsMerchant {
			merchants = append(merchants, user)
		}
	}

	// Generate customer ratings for merchants
	for _, merchant := range merchants {
		customerCount := rand.Intn(8) + 3 // 3-10 customers per merchant

		var totalDelivery, totalQuality, totalTransparency, totalEnvironmental int
		var ratingCount int

		for i := 0; i < customerCount; i++ {
			// Pick random customer (non-merchant)
			var customer models.User
			attempts := 0
			for {
				customer = d.users[rand.Intn(len(d.users))]
				if !customer.IsMerchant && customer.ID != merchant.ID {
					break
				}
				attempts++
				if attempts > 50 { // Prevent infinite loop
					log.Printf("Warning: Could not find non-merchant customer for %s after 50 attempts", merchant.Username)
					break
				}
			}

			// Skip if no valid customer found
			if customer.IsMerchant || customer.ID == merchant.ID {
				continue
			}

			// Generate realistic ratings based on merchant profile
			var delivery, quality, transparency, environmental int

			switch merchant.Username {
			case "bob456": // Organic farmer - high quality
				delivery, quality, transparency, environmental = 8+rand.Intn(2), 9+rand.Intn(2), 8+rand.Intn(2), 9+rand.Intn(2)
			case "david999": // Tech startup - good but not perfect
				delivery, quality, transparency, environmental = 7+rand.Intn(2), 7+rand.Intn(2), 8+rand.Intn(2), 7+rand.Intn(2)
			case "frank777": // Traditional retailer - lower ratings
				delivery, quality, transparency, environmental = 5+rand.Intn(3), 6+rand.Intn(2), 5+rand.Intn(3), 4+rand.Intn(3)
			case "henry111": // Renewable energy - high environmental
				delivery, quality, transparency, environmental = 8+rand.Intn(2), 8+rand.Intn(2), 9+rand.Intn(2), 10
			default:
				delivery, quality, transparency, environmental = 6+rand.Intn(3), 6+rand.Intn(3), 6+rand.Intn(3), 6+rand.Intn(3)
			}

			// Ensure ratings are within 1-10 range
			if delivery > 10 {
				delivery = 10
			}
			if quality > 10 {
				quality = 10
			}
			if transparency > 10 {
				transparency = 10
			}
			if environmental > 10 {
				environmental = 10
			}

			rating := &models.Rating{
				UserID:              customer.ID,
				MerchantID:          merchant.ID,
				DeliveryRating:      delivery,
				QualityRating:       quality,
				TransparencyRating:  transparency,
				EnvironmentalRating: environmental,
				Comments:            d.generateRatingComment(merchant.Username, quality),
				CreatedAt:           time.Now().AddDate(0, 0, -rand.Intn(60)),
			}

			// Create rating using transaction to ensure it's properly committed
			tx := d.db.Begin()
			if err := tx.Create(rating).Error; err != nil {
				tx.Rollback()
				log.Printf("Error creating rating: %v", err)
				continue
			}
			if err := tx.Commit().Error; err != nil {
				log.Printf("Error committing rating: %v", err)
				continue
			}

			log.Printf("Created rating for merchant %s from customer %s", merchant.Username, customer.Username)

			totalDelivery += delivery
			totalQuality += quality
			totalTransparency += transparency
			totalEnvironmental += environmental
			ratingCount++
		}

		if ratingCount > 0 {
			avgDelivery := float64(totalDelivery) / float64(ratingCount)
			avgQuality := float64(totalQuality) / float64(ratingCount)
			avgTransparency := float64(totalTransparency) / float64(ratingCount)
			avgEnvironmental := float64(totalEnvironmental) / float64(ratingCount)

			fmt.Printf("   🏪 %s (%s): %d customers\n", merchant.FirstName, merchant.Username, ratingCount)
			fmt.Printf("      Avg Ratings - Delivery: %.1f, Quality: %.1f, Transparency: %.1f, Environmental: %.1f\n",
				avgDelivery, avgQuality, avgTransparency, avgEnvironmental)
		}
	}

	fmt.Println("   ✅ Merchant activities and ratings generated")

	// Add small delay to ensure all database operations are complete
	time.Sleep(100 * time.Millisecond)
}

func (d *Demo) generateRatingComment(merchantUsername string, qualityRating int) string {
	comments := map[string][]string{
		"bob456": {
			"Excellent organic produce, very fresh!",
			"Love supporting local sustainable farming",
			"Great quality vegetables, will buy again",
		},
		"david999": {
			"Good service, tech solutions work well",
			"Professional and reliable",
			"Innovative approach to business",
		},
		"frank777": {
			"Average service, room for improvement",
			"Traditional approach, could be more transparent",
			"Decent products but not exceptional",
		},
		"henry111": {
			"Outstanding environmental commitment!",
			"Solar installation was perfect",
			"Really cares about sustainability",
		},
	}

	if merchantComments, exists := comments[merchantUsername]; exists {
		return merchantComments[rand.Intn(len(merchantComments))]
	}

	if qualityRating >= 8 {
		return "Good experience overall"
	} else if qualityRating >= 6 {
		return "Average service"
	}
	return "Could be better"
}

func (d *Demo) calculateFairnessScores() {
	fmt.Println("\n📊 Step 5: Calculating PFI★ and TFI★ scores...")

	// Check if we have users to calculate fairness scores for
	if len(d.users) == 0 {
		fmt.Println("   ⚠️  No users found, skipping fairness score calculations")
		return
	}

	for i, user := range d.users {
		// Skip fairness calculations for admin users
		if user.IsAdmin {
			fmt.Printf("   🔧 %s (%s): Admin user - no fairness scoring\n", user.FirstName, user.Username)
			continue
		}

		// Calculate PFI (Personal Fairness Index) using the service method
		err := d.fairnessService.UpdateUserPFI(user.ID)
		if err != nil {
			log.Printf("Error calculating PFI for %s: %v", user.Username, err)
			continue
		}

		// Calculate TFI (Trade Fairness Index) for merchants
		if user.IsMerchant {
			// Check ratings count before TFI calculation
			var ratingCount int64
			d.db.Model(&models.Rating{}).Where("merchant_id = ?", user.ID).Count(&ratingCount)

			err = d.fairnessService.UpdateMerchantTFI(user.ID)
			if err != nil {
				log.Printf("Error calculating TFI for %s: %v", user.Username, err)
			} else {
				log.Printf("TFI calculated for %s with %d ratings", user.Username, ratingCount)
			}
		}

		// Fetch updated user data
		var updatedUser models.User
		if err := d.db.First(&updatedUser, "id = ?", user.ID).Error; err != nil {
			log.Printf("Error fetching updated user %s: %v", user.Username, err)
			continue
		}

		// Update local copy
		d.users[i].PFI = updatedUser.PFI
		d.users[i].TFI = updatedUser.TFI

		pfiCategory := d.getPFICategory(d.users[i].PFI)
		fmt.Printf("   ⭐ %s (%s): PFI★ %d (%s)", user.FirstName, user.Username, d.users[i].PFI, pfiCategory)

		if user.IsMerchant {
			tfiCategory := d.getTFICategory(d.users[i].TFI)
			fmt.Printf(", TFI★ %d (%s)", d.users[i].TFI, tfiCategory)
		}
		fmt.Println()
	}

	fmt.Println("   ✅ Fairness scores calculated")
}

func (d *Demo) getPFICategory(pfi int) string {
	if pfi >= 90 {
		return "Exceptional"
	}
	if pfi >= 80 {
		return "Excellent"
	}
	if pfi >= 70 {
		return "Good"
	}
	if pfi >= 60 {
		return "Average"
	}
	return "Developing"
}

func (d *Demo) getTFICategory(tfi int) string {
	if tfi >= 90 {
		return "Outstanding"
	}
	if tfi >= 80 {
		return "Excellent"
	}
	if tfi >= 70 {
		return "Good"
	}
	if tfi >= 60 {
		return "Fair"
	}
	return "Needs Improvement"
}

func (d *Demo) generateTransactionHistory() {
	fmt.Println("\n💰 Step 6: Generating realistic transaction history...")

	// Check if we have users to generate transactions for
	if len(d.users) == 0 {
		fmt.Println("   ⚠️  No users found, skipping transaction generation")
		return
	}

	// Generate initial wallet balances
	for _, user := range d.users {
		initialBalance := float64(rand.Intn(500) + 100) // 100-600 FC

		// Create wallet directly
		wallet := &models.Wallet{
			UserID:  user.ID,
			Balance: initialBalance,
		}

		if err := d.db.Create(wallet).Error; err != nil {
			log.Printf("Error creating wallet for %s: %v", user.Username, err)
		} else {
			fmt.Printf("   💳 %s: %.2f FC initial balance\n", user.FirstName, wallet.Balance)
		}
	}

	// Generate diverse transactions over the past 3 months
	transactionCount := 0
	totalVolume := 0.0

	// Monthly issuance rewards (based on PFI)
	for _, user := range d.users {
		for month := 2; month >= 0; month-- {
			// Higher PFI users get more monthly rewards
			rewardMultiplier := float64(user.PFI)/100.0*0.5 + 0.5 // 0.5x to 1.0x
			baseReward := 50.0
			reward := baseReward * rewardMultiplier

			transaction := &models.Transaction{
				UserID:      user.ID,
				Type:        models.TransactionTypeMonthlyIssuance,
				Amount:      reward,
				Description: "Monthly FairCoin issuance reward",
				Status:      "completed",
				CreatedAt:   time.Now().AddDate(0, -month, -rand.Intn(28)),
			}

			if err := d.db.Create(transaction).Error; err == nil {
				transactionCount++
				totalVolume += reward
			}
		}
	}

	// Fairness rewards (based on community activities)
	for _, user := range d.users {
		rewardCount := user.CommunityService / 10 // 1 reward per 10 hours
		if rewardCount > 5 {
			rewardCount = 5
		} // Max 5 rewards

		for i := 0; i < rewardCount; i++ {
			reward := float64(rand.Intn(30) + 20) // 20-50 FC

			transaction := &models.Transaction{
				UserID:      user.ID,
				Type:        models.TransactionTypeFairnessReward,
				Amount:      reward,
				Description: "Community service fairness reward",
				Status:      "completed",
				CreatedAt:   time.Now().AddDate(0, 0, -rand.Intn(90)),
			}

			if err := d.db.Create(transaction).Error; err == nil {
				transactionCount++
				totalVolume += reward
			}
		}
	}

	// Merchant incentives
	for _, user := range d.users {
		if user.IsMerchant {
			incentiveCount := 2 + rand.Intn(3) // 2-4 incentives per merchant

			for i := 0; i < incentiveCount; i++ {
				// Higher TFI merchants get bigger incentives
				incentiveMultiplier := float64(user.TFI)/100.0*0.5 + 0.5
				baseIncentive := 75.0
				incentive := baseIncentive * incentiveMultiplier

				transaction := &models.Transaction{
					UserID:      user.ID,
					Type:        models.TransactionTypeMerchantIncentive,
					Amount:      incentive,
					Description: "Fair trade merchant incentive",
					Status:      "completed",
					CreatedAt:   time.Now().AddDate(0, 0, -rand.Intn(90)),
				}

				if err := d.db.Create(transaction).Error; err == nil {
					transactionCount++
					totalVolume += incentive
				}
			}
		}
	}

	// Peer-to-peer transfers
	transferCount := 150 + rand.Intn(100) // 150-250 transfers
	for i := 0; i < transferCount; i++ {
		fromUser := d.users[rand.Intn(len(d.users))]
		toUser := d.users[rand.Intn(len(d.users))]

		// Avoid self-transfers
		for fromUser.ID == toUser.ID {
			toUser = d.users[rand.Intn(len(d.users))]
		}

		amount := float64(rand.Intn(200) + 10) // 10-210 FC
		fee := amount * 0.001                  // 0.1% fee

		transaction := &models.Transaction{
			UserID:      fromUser.ID,
			ToUserID:    &toUser.ID,
			Type:        models.TransactionTypeTransfer,
			Amount:      amount,
			Fee:         fee,
			Description: "P2P transfer",
			Status:      "completed",
			CreatedAt:   time.Now().AddDate(0, 0, -rand.Intn(90)),
		}

		if err := d.db.Create(transaction).Error; err == nil {
			transactionCount++
			totalVolume += amount
		}
	}

	fmt.Printf("   📈 Generated %d transactions with %.2f FC total volume\n", transactionCount, totalVolume)
	fmt.Println("   ✅ Transaction history generated")
}

func (d *Demo) createGovernanceProposals() {
	fmt.Println("\n🗳️  Step 7: Creating governance proposals...")

	// Check if we have users to create governance proposals for
	if len(d.users) == 0 {
		fmt.Println("   ⚠️  No users found, skipping governance proposals creation")
		return
	}

	proposals := []struct {
		proposerIndex int
		title         string
		description   string
		type_         models.ProposalType
		status        models.ProposalStatus
	}{
		{
			0, // Alice
			"Increase Environmental Action Rewards",
			"Proposal to increase fairness rewards for verified environmental activities by 25% to encourage more sustainable practices in our community.",
			models.ProposalTypeCommunity,
			models.ProposalStatusActive,
		},
		{
			2, // Carol
			"Community Mediation Fund",
			"Establish a dedicated fund to support community dispute resolution services, funded by 2% of monthly issuance.",
			models.ProposalTypeMonetaryPolicy,
			models.ProposalStatusActive,
		},
		{
			6, // Grace
			"Senior Community Member Recognition",
			"Create special recognition and rewards for community members over 65 who contribute their wisdom and experience.",
			models.ProposalTypeCommunity,
			models.ProposalStatusPassed,
		},
		{
			3, // David
			"Digital Literacy Initiative Funding",
			"Allocate resources for digital literacy programs to help all community members participate in the digital economy.",
			models.ProposalTypeTechnical,
			models.ProposalStatusActive,
		},
	}

	for _, prop := range proposals {
		if prop.proposerIndex < len(d.users) {
			proposer := d.users[prop.proposerIndex]

			startTime := time.Now().AddDate(0, 0, -rand.Intn(30))
			endTime := startTime.AddDate(0, 0, 14) // 2-week voting period

			proposal := &models.Proposal{
				ProposerID:  proposer.ID,
				Title:       prop.title,
				Description: prop.description,
				Type:        prop.type_,
				Status:      prop.status,
				StartTime:   startTime,
				EndTime:     endTime,
				CreatedAt:   startTime,
			}

			if err := d.db.Create(proposal).Error; err != nil {
				log.Printf("Error creating proposal: %v", err)
				continue
			}

			// Generate votes for the proposal
			d.generateVotes(proposal)

			fmt.Printf("   📋 \"%s\" by %s (%s)\n", prop.title, proposer.FirstName, prop.status)
		}
	}

	fmt.Println("   ✅ Governance proposals created")
}

func (d *Demo) generateVotes(proposal *models.Proposal) {
	// Random number of voters (50-80% participation)
	voterCount := len(d.users) * (50 + rand.Intn(30)) / 100

	for i := 0; i < voterCount; i++ {
		voter := d.users[rand.Intn(len(d.users))]

		// Calculate voting power (60% wallet balance + 40% PFI)
		// For demo purposes, assume equal wallet balances
		totalSupply := 10000.0 // Estimated total FairCoin supply
		walletBalance := 200.0 // Average wallet balance
		votingPower := voter.CalculateVotingPower(walletBalance, totalSupply)

		// Vote based on proposal type and voter profile
		vote := d.determineVote(proposal, voter)

		voteRecord := &models.Vote{
			UserID:      voter.ID,
			ProposalID:  proposal.ID,
			Vote:        vote,
			VotingPower: votingPower,
			CreatedAt:   proposal.StartTime.AddDate(0, 0, rand.Intn(14)),
		}

		if err := d.db.Create(voteRecord).Error; err != nil {
			continue // Skip duplicate votes (same user voting twice)
		}

		// Update proposal vote counts
		if vote {
			proposal.VotesFor++
		} else {
			proposal.VotesAgainst++
		}
		proposal.VotingPower += votingPower
	}

	// Update proposal in database
	d.db.Save(proposal)
}

func (d *Demo) determineVote(proposal *models.Proposal, voter models.User) bool {
	// Simple voting logic based on user characteristics
	voteChance := 0.5 // Base 50% chance to vote yes

	// High PFI users more likely to vote for community proposals
	if proposal.Type == models.ProposalTypeCommunity && voter.PFI >= 80 {
		voteChance += 0.3
	}

	// Merchants more interested in monetary policy
	if proposal.Type == models.ProposalTypeMonetaryPolicy && voter.IsMerchant {
		voteChance += 0.2
	}

	// Environmental activists like environmental proposals
	if voter.Username == "alice123" && proposal.Title == "Increase Environmental Action Rewards" {
		voteChance = 0.9
	}

	return rand.Float64() < voteChance
}

func (d *Demo) displayResults() {
	fmt.Println("\n📊 Step 8: Final Demonstration Results")
	fmt.Println("=====================================")

	// User Statistics
	fmt.Println("\n👥 USER PROFILES & FAIRNESS SCORES:")
	fmt.Println("-----------------------------------")
	if len(d.users) == 0 {
		fmt.Println("⚠️  No users were created during this demo run")
	} else {
		for _, user := range d.users {
			status := "Regular User"
			if user.IsAdmin {
				status = "Administrator"
			} else if user.IsMerchant {
				status = "Merchant"
			}

			fmt.Printf("🧑 %s %s (%s) - %s\n", user.FirstName, user.LastName, user.Username, status)

			if user.IsAdmin {
				fmt.Printf("   System Administrator - No fairness scoring\n")
			} else {
				fmt.Printf("   PFI★: %d (%s)", user.PFI, d.getPFICategory(user.PFI))
				if user.IsMerchant {
					fmt.Printf(", TFI★: %d (%s)", user.TFI, d.getTFICategory(user.TFI))
				}
				fmt.Printf(", Community Service: %d hours\n", user.CommunityService)
			}
		}
	}

	// Transaction Statistics
	fmt.Println("\n💰 TRANSACTION SUMMARY:")
	fmt.Println("----------------------")

	var stats []struct {
		Type   models.TransactionType
		Count  int64
		Volume float64
	}

	types := []models.TransactionType{
		models.TransactionTypeTransfer,
		models.TransactionTypeFairnessReward,
		models.TransactionTypeMerchantIncentive,
		models.TransactionTypeMonthlyIssuance,
	}

	for _, txType := range types {
		var count int64
		var volume float64
		d.db.Model(&models.Transaction{}).Where("type = ?", txType).Count(&count)
		d.db.Model(&models.Transaction{}).Where("type = ?", txType).Select("COALESCE(SUM(amount), 0)").Row().Scan(&volume)

		stats = append(stats, struct {
			Type   models.TransactionType
			Count  int64
			Volume float64
		}{txType, count, volume})
	}

	for _, stat := range stats {
		fmt.Printf("📈 %s: %d transactions, %.2f FC\n", stat.Type, stat.Count, stat.Volume)
	}

	// Governance Statistics
	fmt.Println("\n🗳️  GOVERNANCE ACTIVITY:")
	fmt.Println("----------------------")

	var proposalCount int64
	d.db.Model(&models.Proposal{}).Count(&proposalCount)

	var voteCount int64
	d.db.Model(&models.Vote{}).Count(&voteCount)

	fmt.Printf("📋 Total Proposals: %d\n", proposalCount)
	fmt.Printf("🗳️  Total Votes Cast: %d\n", voteCount)

	// Community Engagement
	fmt.Println("\n🤝 COMMUNITY ENGAGEMENT:")
	fmt.Println("------------------------")

	var attestationCount int64
	d.db.Model(&models.Attestation{}).Count(&attestationCount)

	var ratingCount int64
	d.db.Model(&models.Rating{}).Count(&ratingCount)

	fmt.Printf("⭐ Peer Attestations: %d\n", attestationCount)
	fmt.Printf("🛍️  Merchant Ratings: %d\n", ratingCount)

	// Top Contributors
	fmt.Println("\n🏆 TOP CONTRIBUTORS:")
	fmt.Println("-------------------")

	// Sort users by PFI
	sortedUsers := make([]models.User, len(d.users))
	copy(sortedUsers, d.users)

	// Simple bubble sort by PFI (descending)
	for i := 0; i < len(sortedUsers)-1; i++ {
		for j := 0; j < len(sortedUsers)-i-1; j++ {
			if sortedUsers[j].PFI < sortedUsers[j+1].PFI {
				sortedUsers[j], sortedUsers[j+1] = sortedUsers[j+1], sortedUsers[j]
			}
		}
	}

	fmt.Println("🥇 Highest PFI★ Scores:")
	for i := 0; i < 3 && i < len(sortedUsers); i++ {
		user := sortedUsers[i]
		fmt.Printf("   %d. %s %s: %d PFI★ (%s)\n", i+1, user.FirstName, user.LastName, user.PFI, d.getPFICategory(user.PFI))
	}

	// Top merchants by TFI
	merchantUsers := []models.User{}
	for _, user := range d.users {
		if user.IsMerchant {
			merchantUsers = append(merchantUsers, user)
		}
	}

	if len(merchantUsers) > 0 {
		// Sort merchants by TFI
		for i := 0; i < len(merchantUsers)-1; i++ {
			for j := 0; j < len(merchantUsers)-i-1; j++ {
				if merchantUsers[j].TFI < merchantUsers[j+1].TFI {
					merchantUsers[j], merchantUsers[j+1] = merchantUsers[j+1], merchantUsers[j]
				}
			}
		}

		fmt.Println("\n🏪 Top Merchants by TFI★:")
		for i := 0; i < len(merchantUsers) && i < 3; i++ {
			user := merchantUsers[i]
			fmt.Printf("   %d. %s %s: %d TFI★ (%s)\n", i+1, user.FirstName, user.LastName, user.TFI, d.getTFICategory(user.TFI))
		}
	}

	fmt.Println("\n🎯 DEMONSTRATION HIGHLIGHTS:")
	fmt.Println("---------------------------")
	fmt.Println("✅ Personal Fairness Index (PFI★) calculated based on:")
	fmt.Println("   • Community service hours")
	fmt.Println("   • Peer attestations and ratings")
	fmt.Println("   • Dispute resolution participation")
	fmt.Println("   • Environmental and social activities")

	fmt.Println("\n✅ Trade Fairness Index (TFI★) calculated for merchants based on:")
	fmt.Println("   • Customer satisfaction ratings")
	fmt.Println("   • Delivery reliability")
	fmt.Println("   • Product/service quality")
	fmt.Println("   • Transparency in business practices")
	fmt.Println("   • Environmental responsibility")

	fmt.Println("\n✅ FairCoin ecosystem features demonstrated:")
	fmt.Println("   • Merit-based rewards system")
	fmt.Println("   • Democratic governance with PFI-weighted voting")
	fmt.Println("   • Incentives for community contribution")
	fmt.Println("   • Fair trade merchant support")
	fmt.Println("   • Transparent and accountable transactions")
}
