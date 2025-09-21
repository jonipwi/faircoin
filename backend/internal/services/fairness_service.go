package services

import (
	"faircoin/internal/models"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// FairnessService handles PFI and TFI calculations
type FairnessService struct {
	db *gorm.DB
}

// NewFairnessService creates a new fairness service
func NewFairnessService(db *gorm.DB) *FairnessService {
	return &FairnessService{db: db}
}

// CreateAttestation creates a new attestation for PFI calculation
func (s *FairnessService) CreateAttestation(userID, attesterID uuid.UUID, attestationType string, value int, description string) (*models.Attestation, error) {
	// Verify attester has sufficient PFI to make attestations
	var attester models.User
	if err := s.db.First(&attester, "id = ?", attesterID).Error; err != nil {
		return nil, fmt.Errorf("attester not found: %w", err)
	}

	if attester.PFI < 30 {
		return nil, fmt.Errorf("attester PFI too low to provide attestations")
	}

	// Check for duplicate attestations
	var existingCount int64
	s.db.Model(&models.Attestation{}).
		Where("user_id = ? AND attester_id = ? AND type = ?", userID, attesterID, attestationType).
		Count(&existingCount)

	if existingCount > 0 {
		return nil, fmt.Errorf("attestation already exists")
	}

	attestation := &models.Attestation{
		UserID:      userID,
		AttesterID:  attesterID,
		Type:        attestationType,
		Value:       value,
		Description: description,
		Verified:    false, // Requires community verification
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(attestation).Error; err != nil {
		return nil, fmt.Errorf("failed to create attestation: %w", err)
	}

	// Auto-verify if attester has very high PFI
	if attester.PFI >= 80 {
		attestation.Verified = true
		s.db.Save(attestation)
	}

	// Recalculate PFI for the user (synchronously to avoid database locking)
	if err := s.UpdateUserPFI(userID); err != nil {
		// Log error but don't fail the attestation creation
		fmt.Printf("Warning: Failed to update PFI for user %s: %v\n", userID, err)
	}

	return attestation, nil
}

// UpdateUserPFI recalculates and updates a user's PFI score
func (s *FairnessService) UpdateUserPFI(userID uuid.UUID) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Get all verified attestations
	var attestations []models.Attestation
	s.db.Where("user_id = ? AND verified = ?", userID, true).Find(&attestations)

	// Calculate PFI based on different factors
	pfi := s.calculatePFIScore(&user, attestations)

	// Update user PFI
	user.PFI = int(math.Min(100, math.Max(0, pfi)))
	return s.db.Save(&user).Error
}

// calculatePFIScore calculates PFI based on various factors
func (s *FairnessService) calculatePFIScore(user *models.User, attestations []models.Attestation) float64 {
	basePFI := 10.0 // Starting score

	// Community service hours (up to 30 points)
	servicePoints := math.Min(30, float64(user.CommunityService)*0.5)

	// Peer attestations (up to 40 points)
	var attestationPoints float64
	for _, att := range attestations {
		switch att.Type {
		case "community_service":
			attestationPoints += float64(att.Value) * 2.0
		case "dispute_resolution":
			attestationPoints += float64(att.Value) * 1.5
		case "peer_rating":
			attestationPoints += float64(att.Value) * 0.8
		case "identity_verification":
			attestationPoints += float64(att.Value) * 1.0
		}
	}
	attestationPoints = math.Min(40, attestationPoints)

	// Account age bonus (up to 10 points)
	accountAge := time.Since(user.CreatedAt).Hours() / 24 / 30 // months
	agePoints := math.Min(10, accountAge*0.5)

	// Transaction behavior (up to 20 points)
	var transactionPoints float64
	var transactionCount int64
	s.db.Model(&models.Transaction{}).Where("user_id = ?", user.ID).Count(&transactionCount)

	if transactionCount > 0 {
		// Points for regular transactions
		transactionPoints += math.Min(10, float64(transactionCount)*0.1)

		// Check for dispute-free transactions
		var disputeCount int64
		// This would require a disputes table - for now assume low disputes
		disputeRate := float64(disputeCount) / float64(transactionCount)
		if disputeRate < 0.1 {
			transactionPoints += 10
		}
	}

	totalPFI := basePFI + servicePoints + attestationPoints + agePoints + transactionPoints
	return math.Min(100, totalPFI)
}

// CreateRating creates a merchant rating for TFI calculation
func (s *FairnessService) CreateRating(userID, merchantID uuid.UUID, transactionID *uuid.UUID,
	deliveryRating, qualityRating, transparencyRating, environmentalRating int, comments string) (*models.Rating, error) {

	// Verify the merchant exists and is actually a merchant
	var merchant models.User
	if err := s.db.First(&merchant, "id = ? AND is_merchant = ?", merchantID, true).Error; err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	// Check if user has already rated this merchant recently
	var existingCount int64
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	s.db.Model(&models.Rating{}).
		Where("user_id = ? AND merchant_id = ? AND created_at > ?", userID, merchantID, oneWeekAgo).
		Count(&existingCount)

	if existingCount > 0 {
		return nil, fmt.Errorf("you have already rated this merchant recently")
	}

	rating := &models.Rating{
		UserID:              userID,
		MerchantID:          merchantID,
		TransactionID:       transactionID,
		DeliveryRating:      deliveryRating,
		QualityRating:       qualityRating,
		TransparencyRating:  transparencyRating,
		EnvironmentalRating: environmentalRating,
		Comments:            comments,
		CreatedAt:           time.Now(),
	}

	if err := s.db.Create(rating).Error; err != nil {
		return nil, fmt.Errorf("failed to create rating: %w", err)
	}

	// Recalculate TFI for the merchant (synchronously to avoid database locking)
	if err := s.UpdateMerchantTFI(merchantID); err != nil {
		// Log error but don't fail the rating creation
		fmt.Printf("Warning: Failed to update TFI for merchant %s: %v\n", merchantID, err)
	}

	return rating, nil
}

// UpdateMerchantTFI recalculates and updates a merchant's TFI score
func (s *FairnessService) UpdateMerchantTFI(merchantID uuid.UUID) error {
	// Use transaction to prevent database locking issues
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var merchant models.User
	if err := tx.First(&merchant, "id = ? AND is_merchant = ?", merchantID, true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("merchant not found: %w", err)
	}

	// Get all ratings
	var ratings []models.Rating
	if err := tx.Where("merchant_id = ?", merchantID).Find(&ratings).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get ratings: %w", err)
	}

	if len(ratings) == 0 {
		// New merchant starts with base TFI if not already set
		if merchant.TFI == 0 {
			merchant.TFI = 30
		}
	} else {
		// Calculate TFI based on ratings
		tfi := s.calculateTFIScore(&merchant, ratings)
		merchant.TFI = int(math.Min(100, math.Max(30, tfi))) // Minimum TFI of 30 for merchants with ratings
	}

	if err := tx.Save(&merchant).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save merchant: %w", err)
	}

	return tx.Commit().Error
}

// calculateTFIScore calculates TFI based on merchant ratings
func (s *FairnessService) calculateTFIScore(merchant *models.User, ratings []models.Rating) float64 {
	if len(ratings) == 0 {
		return 30.0 // Base score for new merchants
	}

	var totalDelivery, totalQuality, totalTransparency, totalEnvironmental float64

	for _, rating := range ratings {
		totalDelivery += float64(rating.DeliveryRating)
		totalQuality += float64(rating.QualityRating)
		totalTransparency += float64(rating.TransparencyRating)
		totalEnvironmental += float64(rating.EnvironmentalRating)
	}

	count := float64(len(ratings))
	avgDelivery := totalDelivery / count
	avgQuality := totalQuality / count
	avgTransparency := totalTransparency / count
	avgEnvironmental := totalEnvironmental / count

	// Weighted TFI calculation
	tfi := (avgDelivery*0.3 + avgQuality*0.3 + avgTransparency*0.25 + avgEnvironmental*0.15) * 10

	// Bonus for having many ratings (trust factor)
	if count > 10 {
		tfi += math.Min(10, (count-10)*0.1)
	}

	return tfi
}

// UpdateAllScores updates PFI and TFI scores for all users (called periodically)
func (s *FairnessService) UpdateAllScores() error {
	// Update PFI for all users
	var users []models.User
	s.db.Find(&users)

	for _, user := range users {
		if err := s.UpdateUserPFI(user.ID); err != nil {
			// Log error but continue
			fmt.Printf("Error updating PFI for user %s: %v\n", user.Username, err)
		}
	}

	// Update TFI for all merchants
	var merchants []models.User
	s.db.Where("is_merchant = ?", true).Find(&merchants)

	for _, merchant := range merchants {
		if err := s.UpdateMerchantTFI(merchant.ID); err != nil {
			// Log error but continue
			fmt.Printf("Error updating TFI for merchant %s: %v\n", merchant.Username, err)
		}
	}

	return nil
}

// GetUserPFIBreakdown returns detailed PFI calculation breakdown
func (s *FairnessService) GetUserPFIBreakdown(userID uuid.UUID) (map[string]interface{}, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var attestations []models.Attestation
	s.db.Where("user_id = ? AND verified = ?", userID, true).Find(&attestations)

	breakdown := make(map[string]interface{})
	breakdown["current_pfi"] = user.PFI
	breakdown["community_service_hours"] = user.CommunityService
	breakdown["total_attestations"] = len(attestations)
	breakdown["account_age_days"] = int(time.Since(user.CreatedAt).Hours() / 24)

	// Count attestations by type
	attestationTypes := make(map[string]int)
	for _, att := range attestations {
		attestationTypes[att.Type]++
	}
	breakdown["attestation_breakdown"] = attestationTypes

	return breakdown, nil
}

// GetMerchantTFIBreakdown returns detailed TFI calculation breakdown
func (s *FairnessService) GetMerchantTFIBreakdown(merchantID uuid.UUID) (map[string]interface{}, error) {
	var merchant models.User
	if err := s.db.First(&merchant, "id = ? AND is_merchant = ?", merchantID, true).Error; err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	var ratings []models.Rating
	s.db.Where("merchant_id = ?", merchantID).Find(&ratings)

	breakdown := make(map[string]interface{})
	breakdown["current_tfi"] = merchant.TFI
	breakdown["total_ratings"] = len(ratings)

	if len(ratings) > 0 {
		var totalDelivery, totalQuality, totalTransparency, totalEnvironmental float64
		for _, rating := range ratings {
			totalDelivery += float64(rating.DeliveryRating)
			totalQuality += float64(rating.QualityRating)
			totalTransparency += float64(rating.TransparencyRating)
			totalEnvironmental += float64(rating.EnvironmentalRating)
		}

		count := float64(len(ratings))
		breakdown["avg_delivery_rating"] = math.Round((totalDelivery/count)*100) / 100
		breakdown["avg_quality_rating"] = math.Round((totalQuality/count)*100) / 100
		breakdown["avg_transparency_rating"] = math.Round((totalTransparency/count)*100) / 100
		breakdown["avg_environmental_rating"] = math.Round((totalEnvironmental/count)*100) / 100
	}

	return breakdown, nil
}
