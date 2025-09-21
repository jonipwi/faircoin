package services

import (
	"faircoin/internal/models"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// GovernanceService handles governance and voting operations
type GovernanceService struct {
	db *gorm.DB
}

// NewGovernanceService creates a new governance service
func NewGovernanceService(db *gorm.DB) *GovernanceService {
	return &GovernanceService{db: db}
}

// GetDB returns the database connection
func (s *GovernanceService) GetDB() *gorm.DB {
	return s.db
}

// CreateProposal creates a new governance proposal
func (s *GovernanceService) CreateProposal(proposerID uuid.UUID, title, description string, proposalType models.ProposalType) (*models.Proposal, error) {
	// Check if proposer has sufficient PFI
	var proposer models.User
	if err := s.db.First(&proposer, "id = ?", proposerID).Error; err != nil {
		return nil, fmt.Errorf("proposer not found: %w", err)
	}

	if proposer.PFI < 50 { // Minimum PFI for proposals
		return nil, fmt.Errorf("insufficient PFI to create proposals (minimum: 50, current: %d)", proposer.PFI)
	}

	proposal := &models.Proposal{
		ProposerID:  proposerID,
		Title:       title,
		Description: description,
		Type:        proposalType,
		Status:      models.ProposalStatusActive,
		StartTime:   time.Now(),
		EndTime:     time.Now().AddDate(0, 0, 7), // 7 days voting period
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(proposal).Error; err != nil {
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}

	return proposal, nil
}

// VoteOnProposal allows a user to vote on a proposal
func (s *GovernanceService) VoteOnProposal(userID, proposalID uuid.UUID, vote bool) error {
	// Get proposal
	var proposal models.Proposal
	if err := s.db.First(&proposal, "id = ?", proposalID).Error; err != nil {
		return fmt.Errorf("proposal not found: %w", err)
	}

	// Check if proposal is still active
	if proposal.Status != models.ProposalStatusActive {
		return fmt.Errorf("proposal is not active")
	}

	// Check if voting period is still open
	if time.Now().After(proposal.EndTime) {
		return fmt.Errorf("voting period has ended")
	}

	// Check if user has already voted
	var existingVote models.Vote
	if err := s.db.Where("user_id = ? AND proposal_id = ?", userID, proposalID).First(&existingVote).Error; err == nil {
		return fmt.Errorf("user has already voted on this proposal")
	}

	// Get user and calculate voting power
	var user models.User
	if err := s.db.Preload("Wallet").First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Calculate total supply for voting power calculation
	var totalSupply struct {
		Total float64
	}
	s.db.Model(&models.Wallet{}).Select("SUM(balance) as total").Scan(&totalSupply)

	// Get user's wallet balance
	var wallet models.Wallet
	if err := s.db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	votingPower := user.CalculateVotingPower(wallet.Balance, totalSupply.Total)

	// Create vote record
	voteRecord := &models.Vote{
		UserID:      userID,
		ProposalID:  proposalID,
		Vote:        vote,
		VotingPower: votingPower,
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(voteRecord).Error; err != nil {
		return fmt.Errorf("failed to record vote: %w", err)
	}

	// Update proposal vote counts
	if vote {
		proposal.VotesFor++
	} else {
		proposal.VotesAgainst++
	}
	proposal.VotingPower += votingPower

	return s.db.Save(&proposal).Error
}

// GetActiveProposals returns all active proposals
func (s *GovernanceService) GetActiveProposals() ([]models.Proposal, error) {
	var proposals []models.Proposal
	err := s.db.Preload("Proposer").Where("status = ?", models.ProposalStatusActive).
		Order("created_at DESC").Find(&proposals).Error
	return proposals, err
}

// GetProposalByID returns a specific proposal with votes
func (s *GovernanceService) GetProposalByID(proposalID uuid.UUID) (*models.Proposal, error) {
	var proposal models.Proposal
	err := s.db.Preload("Proposer").Preload("Votes").Preload("Votes.User").
		First(&proposal, "id = ?", proposalID).Error
	return &proposal, err
}

// ProcessExpiredProposals checks and processes expired proposals
func (s *GovernanceService) ProcessExpiredProposals() error {
	var expiredProposals []models.Proposal
	now := time.Now()

	if err := s.db.Where("status = ? AND end_time < ?", models.ProposalStatusActive, now).
		Find(&expiredProposals).Error; err != nil {
		return err
	}

	for _, proposal := range expiredProposals {
		// Determine if proposal passed
		// Simple majority with minimum participation
		totalVotingPower := proposal.VotingPower
		minParticipation := 0.1 // 10% minimum participation

		if totalVotingPower >= minParticipation {
			if float64(proposal.VotesFor) > float64(proposal.VotesAgainst) {
				proposal.Status = models.ProposalStatusPassed
			} else {
				proposal.Status = models.ProposalStatusRejected
			}
		} else {
			proposal.Status = models.ProposalStatusRejected // Not enough participation
		}

		s.db.Save(&proposal)
	}

	return nil
}

// GetCouncilMembers returns the current community council members
func (s *GovernanceService) GetCouncilMembers() ([]models.User, error) {
	// Council members are the top PFI users who are active
	var councilMembers []models.User
	err := s.db.Where("pfi >= ? AND is_verified = ?", 70, true).
		Order("pfi DESC").Limit(7).Find(&councilMembers).Error
	return councilMembers, err
}

// MonetaryService handles monetary policy and issuance
type MonetaryService struct {
	db *gorm.DB
}

// NewMonetaryService creates a new monetary service
func NewMonetaryService(db *gorm.DB) *MonetaryService {
	return &MonetaryService{db: db}
}

// GetDB returns the database connection
func (s *MonetaryService) GetDB() *gorm.DB {
	return s.db
}

// ProcessMonthlyIssuance processes the monthly FairCoin issuance
func (s *MonetaryService) ProcessMonthlyIssuance() error {
	currentMonth := time.Now().Format("2006-01")

	// Check if issuance already processed for this month
	var existingPolicy models.MonetaryPolicy
	if err := s.db.Where("month = ?", currentMonth).First(&existingPolicy).Error; err == nil {
		return nil // Already processed
	}

	// Calculate factors for issuance
	baseIssuance := 1000.0 // Base monthly issuance

	// Activity Factor: based on transaction volume
	activityFactor := s.calculateActivityFactor()

	// Fairness Factor: based on average community PFI
	fairnessFactor := s.calculateFairnessFactor()

	// Calculate total issuance
	totalIssuance := baseIssuance * activityFactor * fairnessFactor

	// Apply maximum growth rate cap (2% per month)
	var totalSupply struct {
		Total float64
	}
	s.db.Model(&models.Wallet{}).Select("SUM(balance) as total").Scan(&totalSupply)

	maxIssuance := totalSupply.Total * 0.02 // 2% monthly cap
	if totalIssuance > maxIssuance {
		totalIssuance = maxIssuance
	}

	// Distribute the new issuance
	if err := s.distributeIssuance(totalIssuance); err != nil {
		return fmt.Errorf("failed to distribute issuance: %w", err)
	}

	// Record monetary policy
	policy := &models.MonetaryPolicy{
		Month:             currentMonth,
		BaseIssuance:      baseIssuance,
		ActivityFactor:    activityFactor,
		FairnessFactor:    fairnessFactor,
		TotalIssuance:     totalIssuance,
		CirculatingSupply: totalSupply.Total + totalIssuance,
		CreatedAt:         time.Now(),
	}

	// Get current stats
	var avgPFI struct {
		Average float64
	}
	s.db.Model(&models.User{}).Select("AVG(pfi) as average").Scan(&avgPFI)
	policy.AveragePFI = avgPFI.Average

	var transactionCount int64
	s.db.Model(&models.Transaction{}).Count(&transactionCount)
	policy.TotalTransactions = int(transactionCount)

	return s.db.Create(policy).Error
}

// calculateActivityFactor calculates the activity factor based on transaction volume
func (s *MonetaryService) calculateActivityFactor() float64 {
	// Get transaction volume for the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var volume struct {
		Total float64
		Count int64
	}
	s.db.Model(&models.Transaction{}).
		Where("created_at > ? AND type = ?", thirtyDaysAgo, models.TransactionTypeTransfer).
		Select("SUM(amount) as total, COUNT(*) as count").Scan(&volume)

	// Baseline activity: 1000 FC volume per month
	baselineVolume := 1000.0
	activityFactor := volume.Total / baselineVolume

	// Cap between 0.5 and 1.5
	return math.Max(0.5, math.Min(1.5, activityFactor))
}

// calculateFairnessFactor calculates the fairness factor based on average PFI
func (s *MonetaryService) calculateFairnessFactor() float64 {
	var avgPFI struct {
		Average float64
	}
	s.db.Model(&models.User{}).Select("AVG(pfi) as average").Scan(&avgPFI)

	// Fairness factor: higher community PFI = more issuance
	// Scale: PFI 50 = factor 1.0, PFI 75 = factor 1.25, PFI 25 = factor 0.75
	fairnessFactor := 0.5 + (avgPFI.Average / 100.0)

	// Cap between 0.5 and 1.5
	return math.Max(0.5, math.Min(1.5, fairnessFactor))
}

// distributeIssuance distributes newly minted FairCoins according to policy
func (s *MonetaryService) distributeIssuance(totalIssuance float64) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Distribution: 50% liquidity, 25% fairness rewards, 15% merchant incentives, 10% maintenance
	liquidityAmount := totalIssuance * 0.5
	fairnessAmount := totalIssuance * 0.25
	merchantAmount := totalIssuance * 0.15
	maintenanceAmount := totalIssuance * 0.1

	// 1. Distribute liquidity to active users
	if err := s.distributeLiquidity(tx, liquidityAmount); err != nil {
		tx.Rollback()
		return err
	}

	// 2. Distribute fairness rewards to high-PFI users
	if err := s.distributeFairnessRewards(tx, fairnessAmount); err != nil {
		tx.Rollback()
		return err
	}

	// 3. Distribute merchant incentives to high-TFI merchants
	if err := s.distributeMerchantIncentives(tx, merchantAmount); err != nil {
		tx.Rollback()
		return err
	}

	// 4. Maintenance fund (could be sent to a treasury wallet)
	// For now, just create a record
	maintenanceTransaction := &models.Transaction{
		Type:        models.TransactionTypeMonthlyIssuance,
		Amount:      maintenanceAmount,
		Description: "Monthly maintenance fund allocation",
		Status:      "completed",
		CreatedAt:   time.Now(),
	}
	if err := tx.Create(maintenanceTransaction).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// distributeLiquidity distributes liquidity allocation to active users
func (s *MonetaryService) distributeLiquidity(tx *gorm.DB, amount float64) error {
	// Get active users (had transactions in last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var activeUserIDs []uuid.UUID
	tx.Model(&models.Transaction{}).Where("created_at > ?", thirtyDaysAgo).
		Select("DISTINCT user_id").Pluck("user_id", &activeUserIDs)

	if len(activeUserIDs) == 0 {
		return nil
	}

	// Equal distribution among active users
	amountPerUser := amount / float64(len(activeUserIDs))

	for _, userID := range activeUserIDs {
		// Update wallet balance
		if err := tx.Model(&models.Wallet{}).Where("user_id = ?", userID).
			Update("balance", gorm.Expr("balance + ?", amountPerUser)).Error; err != nil {
			return err
		}

		// Create transaction record
		transaction := &models.Transaction{
			UserID:      userID,
			Type:        models.TransactionTypeMonthlyIssuance,
			Amount:      amountPerUser,
			Description: "Monthly liquidity distribution",
			Status:      "completed",
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}
	}

	return nil
}

// distributeFairnessRewards distributes fairness rewards based on PFI
func (s *MonetaryService) distributeFairnessRewards(tx *gorm.DB, amount float64) error {
	// Get users with PFI >= 50
	var users []models.User
	tx.Where("pfi >= ?", 50).Find(&users)

	if len(users) == 0 {
		return nil
	}

	// Calculate total PFI points
	var totalPFI float64
	for _, user := range users {
		totalPFI += float64(user.PFI)
	}

	// Distribute proportionally to PFI
	for _, user := range users {
		userShare := (float64(user.PFI) / totalPFI) * amount

		// Update wallet balance
		if err := tx.Model(&models.Wallet{}).Where("user_id = ?", user.ID).
			Update("balance", gorm.Expr("balance + ?", userShare)).Error; err != nil {
			return err
		}

		// Create transaction record
		transaction := &models.Transaction{
			UserID:      user.ID,
			Type:        models.TransactionTypeFairnessReward,
			Amount:      userShare,
			Description: fmt.Sprintf("Monthly fairness reward (PFI: %d)", user.PFI),
			Status:      "completed",
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}
	}

	return nil
}

// distributeMerchantIncentives distributes merchant incentives based on TFI
func (s *MonetaryService) distributeMerchantIncentives(tx *gorm.DB, amount float64) error {
	// Get merchants with TFI >= 40
	var merchants []models.User
	tx.Where("is_merchant = ? AND tfi >= ?", true, 40).Find(&merchants)

	if len(merchants) == 0 {
		return nil
	}

	// Calculate total TFI points
	var totalTFI float64
	for _, merchant := range merchants {
		totalTFI += float64(merchant.TFI)
	}

	// Distribute proportionally to TFI
	for _, merchant := range merchants {
		merchantShare := (float64(merchant.TFI) / totalTFI) * amount

		// Update wallet balance
		if err := tx.Model(&models.Wallet{}).Where("user_id = ?", merchant.ID).
			Update("balance", gorm.Expr("balance + ?", merchantShare)).Error; err != nil {
			return err
		}

		// Create transaction record
		transaction := &models.Transaction{
			UserID:      merchant.ID,
			Type:        models.TransactionTypeMerchantIncentive,
			Amount:      merchantShare,
			Description: fmt.Sprintf("Monthly merchant incentive (TFI: %d)", merchant.TFI),
			Status:      "completed",
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetCommunityBasketIndex returns the current CBI
func (s *MonetaryService) GetCommunityBasketIndex() (*models.CommunityBasketIndex, error) {
	var cbi models.CommunityBasketIndex
	err := s.db.Order("created_at DESC").First(&cbi).Error
	if err != nil {
		// Create initial CBI if none exists
		cbi = models.CommunityBasketIndex{
			Value:        100.0, // Base value
			FoodIndex:    100.0,
			EnergyIndex:  100.0,
			LaborIndex:   100.0,
			HousingIndex: 100.0,
			CreatedAt:    time.Now(),
		}
		s.db.Create(&cbi)
	}
	return &cbi, nil
}

// UpdateCommunityBasketIndex updates the CBI with new data
func (s *MonetaryService) UpdateCommunityBasketIndex(foodIndex, energyIndex, laborIndex, housingIndex float64) error {
	// Calculate weighted average
	value := (foodIndex*0.4 + energyIndex*0.2 + laborIndex*0.3 + housingIndex*0.1)

	cbi := &models.CommunityBasketIndex{
		Value:        value,
		FoodIndex:    foodIndex,
		EnergyIndex:  energyIndex,
		LaborIndex:   laborIndex,
		HousingIndex: housingIndex,
		CreatedAt:    time.Now(),
	}

	return s.db.Create(cbi).Error
}

// GetMonetaryPolicyHistory returns monetary policy history
func (s *MonetaryService) GetMonetaryPolicyHistory(months int) ([]models.MonetaryPolicy, error) {
	var policies []models.MonetaryPolicy
	err := s.db.Order("created_at DESC").Limit(months).Find(&policies).Error
	return policies, err
}

// GetCurrentMonthStats returns current month statistics
func (s *MonetaryService) GetCurrentMonthStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	currentMonth := time.Now().Format("2006-01")

	// Get current month policy
	var policy models.MonetaryPolicy
	if err := s.db.Where("month = ?", currentMonth).First(&policy).Error; err == nil {
		stats["monthly_issuance"] = policy.TotalIssuance
		stats["activity_factor"] = policy.ActivityFactor
		stats["fairness_factor"] = policy.FairnessFactor
		stats["circulating_supply"] = policy.CirculatingSupply
	}

	// Get current month transactions
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	var monthlyTransactions int64
	s.db.Model(&models.Transaction{}).Where("created_at >= ?", startOfMonth).Count(&monthlyTransactions)
	stats["monthly_transactions"] = monthlyTransactions

	// Get current month transaction volume
	var monthlyVolume struct {
		Total float64
	}
	s.db.Model(&models.Transaction{}).Where("created_at >= ? AND type = ?", startOfMonth, models.TransactionTypeTransfer).
		Select("SUM(amount) as total").Scan(&monthlyVolume)
	stats["monthly_volume"] = monthlyVolume.Total

	return stats, nil
}
