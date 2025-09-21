package services

import (
	"faircoin/internal/models"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// MetricsService handles comprehensive fairness metrics analysis
type MetricsService struct {
	db *gorm.DB
}

// NewMetricsService creates a new metrics service
func NewMetricsService(db *gorm.DB) *MetricsService {
	return &MetricsService{db: db}
}

// CalculatePFIDistribution calculates the distribution of PFI scores
func (s *MetricsService) CalculatePFIDistribution() (map[string]interface{}, error) {
	var excellent, good, average, poor int64
	var totalUsers int64

	// Count users in each PFI range
	s.db.Model(&models.User{}).Where("pfi >= 90").Count(&excellent)
	s.db.Model(&models.User{}).Where("pfi >= 70 AND pfi < 90").Count(&good)
	s.db.Model(&models.User{}).Where("pfi >= 50 AND pfi < 70").Count(&average)
	s.db.Model(&models.User{}).Where("pfi < 50").Count(&poor)
	s.db.Model(&models.User{}).Count(&totalUsers)

	// Calculate percentages
	excellentPct := 0.0
	goodPct := 0.0
	averagePct := 0.0
	poorPct := 0.0

	if totalUsers > 0 {
		excellentPct = math.Round((float64(excellent)/float64(totalUsers))*100*100) / 100
		goodPct = math.Round((float64(good)/float64(totalUsers))*100*100) / 100
		averagePct = math.Round((float64(average)/float64(totalUsers))*100*100) / 100
		poorPct = math.Round((float64(poor)/float64(totalUsers))*100*100) / 100
	}

	return map[string]interface{}{
		"distribution": map[string]interface{}{
			"excellent": map[string]interface{}{
				"count":      excellent,
				"percentage": excellentPct,
			},
			"good": map[string]interface{}{
				"count":      good,
				"percentage": goodPct,
			},
			"average": map[string]interface{}{
				"count":      average,
				"percentage": averagePct,
			},
			"poor": map[string]interface{}{
				"count":      poor,
				"percentage": poorPct,
			},
		},
		"total_users": totalUsers,
	}, nil
}

// CalculateTFIAnalysis calculates TFI analysis including average rating and merchant rankings
func (s *MetricsService) CalculateTFIAnalysis() (map[string]interface{}, error) {
	// Calculate average TFI
	var avgTFI struct {
		Avg   float64
		Count int64
	}

	s.db.Model(&models.User{}).
		Where("is_merchant = ? AND tfi > 0", true).
		Select("AVG(tfi) as avg, COUNT(*) as count").
		Scan(&avgTFI)

	// Get total number of ratings
	var totalRatings int64
	s.db.Model(&models.Rating{}).Count(&totalRatings)

	// Format average rating (TFI is 0-100, but we want to show as 0-5.0)
	averageRating := 0.0
	if avgTFI.Avg > 0 {
		averageRating = math.Round((avgTFI.Avg/20)*100) / 100 // Convert to 5.0 scale
	}

	return map[string]interface{}{
		"average_rating":  averageRating,
		"average_tfi":     math.Round(avgTFI.Avg*100) / 100,
		"total_ratings":   totalRatings,
		"total_merchants": avgTFI.Count,
		"rating_scale":    "5.0",
	}, nil
}

// GetTopMerchants returns the top merchants by TFI score
func (s *MetricsService) GetTopMerchants(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}

	var merchants []models.User
	err := s.db.Where("is_merchant = ? AND tfi > 0", true).
		Order("tfi DESC").
		Limit(limit).
		Find(&merchants).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get top merchants: %w", err)
	}

	var topMerchants []map[string]interface{}
	for i, merchant := range merchants {
		// Get rating count for this merchant
		var ratingCount int64
		s.db.Model(&models.Rating{}).Where("merchant_id = ?", merchant.ID).Count(&ratingCount)

		// Calculate average rating from TFI (convert back to 5.0 scale)
		averageRating := float64(merchant.TFI) / 20

		topMerchants = append(topMerchants, map[string]interface{}{
			"rank":           i + 1,
			"id":             merchant.ID,
			"username":       merchant.Username,
			"first_name":     merchant.FirstName,
			"last_name":      merchant.LastName,
			"tfi":            merchant.TFI,
			"average_rating": math.Round(averageRating*100) / 100,
			"total_ratings":  ratingCount,
		})
	}

	return topMerchants, nil
}

// CalculateCommunityBasketIndex calculates the current CBI and its components
func (s *MetricsService) CalculateCommunityBasketIndex() (map[string]interface{}, error) {
	// Get the latest CBI entry
	var cbi models.CommunityBasketIndex
	err := s.db.Order("created_at DESC").First(&cbi).Error

	if err != nil {
		// Return default values if no CBI data exists
		return map[string]interface{}{
			"current_cbi": 100.0,
			"trend":       "stable",
			"components": map[string]interface{}{
				"food":    100.0,
				"energy":  100.0,
				"labor":   100.0,
				"housing": 100.0,
			},
			"last_updated": time.Now().Format("2006-01-02"),
		}, nil
	}

	// Calculate trend by comparing with previous entry
	trend := "stable"
	var previousCBI models.CommunityBasketIndex
	if err := s.db.Where("created_at < ?", cbi.CreatedAt).
		Order("created_at DESC").
		First(&previousCBI).Error; err == nil {

		diff := cbi.Value - previousCBI.Value
		if diff > 2 {
			trend = "increasing"
		} else if diff < -2 {
			trend = "decreasing"
		}
	}

	return map[string]interface{}{
		"current_cbi": math.Round(cbi.Value*100) / 100,
		"trend":       trend,
		"components": map[string]interface{}{
			"food":    math.Round(cbi.FoodIndex*100) / 100,
			"energy":  math.Round(cbi.EnergyIndex*100) / 100,
			"labor":   math.Round(cbi.LaborIndex*100) / 100,
			"housing": math.Round(cbi.HousingIndex*100) / 100,
		},
		"last_updated": cbi.CreatedAt.Format("2006-01-02"),
	}, nil
}

// UpdateDailyMetrics calculates and stores daily fairness metrics
func (s *MetricsService) UpdateDailyMetrics() error {
	today := time.Now().Format("2006-01-02")

	// Check if metrics for today already exist
	var existingMetrics models.FairnessMetrics
	if err := s.db.Where("date = ?", today).First(&existingMetrics).Error; err == nil {
		// Update existing metrics
		return s.updateExistingMetrics(&existingMetrics)
	}

	// Create new metrics entry
	return s.createNewMetrics(today)
}

// createNewMetrics creates a new daily metrics entry
func (s *MetricsService) createNewMetrics(date string) error {
	// Calculate PFI distribution
	var excellent, good, average, poor int64
	s.db.Model(&models.User{}).Where("pfi >= 90").Count(&excellent)
	s.db.Model(&models.User{}).Where("pfi >= 70 AND pfi < 90").Count(&good)
	s.db.Model(&models.User{}).Where("pfi >= 50 AND pfi < 70").Count(&average)
	s.db.Model(&models.User{}).Where("pfi < 50").Count(&poor)

	// Calculate TFI metrics
	var avgTFI struct {
		Avg   float64
		Count int64
	}
	s.db.Model(&models.User{}).
		Where("is_merchant = ? AND tfi > 0", true).
		Select("AVG(tfi) as avg, COUNT(*) as count").
		Scan(&avgTFI)

	var totalTFIRatings int64
	s.db.Model(&models.Rating{}).Count(&totalTFIRatings)

	// Get CBI data
	var cbi models.CommunityBasketIndex
	cbiValue := 100.0
	cbiFood := 100.0
	cbiEnergy := 100.0
	cbiLabor := 100.0
	cbiHousing := 100.0
	cbiTrend := "stable"

	if err := s.db.Order("created_at DESC").First(&cbi).Error; err == nil {
		cbiValue = cbi.Value
		cbiFood = cbi.FoodIndex
		cbiEnergy = cbi.EnergyIndex
		cbiLabor = cbi.LaborIndex
		cbiHousing = cbi.HousingIndex

		// Calculate trend
		var previousCBI models.CommunityBasketIndex
		if err := s.db.Where("created_at < ?", cbi.CreatedAt).
			Order("created_at DESC").
			First(&previousCBI).Error; err == nil {

			diff := cbi.Value - previousCBI.Value
			if diff > 2 {
				cbiTrend = "increasing"
			} else if diff < -2 {
				cbiTrend = "decreasing"
			}
		}
	}

	// Get total counts
	var totalUsers, totalMerchants, totalTransactions int64
	s.db.Model(&models.User{}).Count(&totalUsers)
	s.db.Model(&models.User{}).Where("is_merchant = ?", true).Count(&totalMerchants)
	s.db.Model(&models.Transaction{}).Count(&totalTransactions)

	// Create metrics entry
	metrics := &models.FairnessMetrics{
		Date:              date,
		PFIExcellentCount: int(excellent),
		PFIGoodCount:      int(good),
		PFIAverageCount:   int(average),
		PFIPoorCount:      int(poor),
		TotalPFIRatings:   int(excellent + good + average + poor),
		AverageTFI:        avgTFI.Avg,
		TotalTFIRatings:   int(totalTFIRatings),
		CBIValue:          cbiValue,
		CBIFoodIndex:      cbiFood,
		CBIEnergyIndex:    cbiEnergy,
		CBILaborIndex:     cbiLabor,
		CBIHousingIndex:   cbiHousing,
		CBITrend:          cbiTrend,
		TotalUsers:        int(totalUsers),
		TotalMerchants:    int(totalMerchants),
		TotalTransactions: int(totalTransactions),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return s.db.Create(metrics).Error
}

// updateExistingMetrics updates an existing metrics entry
func (s *MetricsService) updateExistingMetrics(metrics *models.FairnessMetrics) error {
	// Recalculate all metrics for today
	var excellent, good, average, poor int64
	s.db.Model(&models.User{}).Where("pfi >= 90").Count(&excellent)
	s.db.Model(&models.User{}).Where("pfi >= 70 AND pfi < 90").Count(&good)
	s.db.Model(&models.User{}).Where("pfi >= 50 AND pfi < 70").Count(&average)
	s.db.Model(&models.User{}).Where("pfi < 50").Count(&poor)

	var avgTFI struct {
		Avg   float64
		Count int64
	}
	s.db.Model(&models.User{}).
		Where("is_merchant = ? AND tfi > 0", true).
		Select("AVG(tfi) as avg, COUNT(*) as count").
		Scan(&avgTFI)

	var totalTFIRatings int64
	s.db.Model(&models.Rating{}).Count(&totalTFIRatings)

	// Get latest CBI
	var cbi models.CommunityBasketIndex
	if err := s.db.Order("created_at DESC").First(&cbi).Error; err == nil {
		metrics.CBIValue = cbi.Value
		metrics.CBIFoodIndex = cbi.FoodIndex
		metrics.CBIEnergyIndex = cbi.EnergyIndex
		metrics.CBILaborIndex = cbi.LaborIndex
		metrics.CBIHousingIndex = cbi.HousingIndex
	}

	// Update metrics
	metrics.PFIExcellentCount = int(excellent)
	metrics.PFIGoodCount = int(good)
	metrics.PFIAverageCount = int(average)
	metrics.PFIPoorCount = int(poor)
	metrics.TotalPFIRatings = int(excellent + good + average + poor)
	metrics.AverageTFI = avgTFI.Avg
	metrics.TotalTFIRatings = int(totalTFIRatings)
	metrics.UpdatedAt = time.Now()

	// Update total counts
	var totalUsers, totalMerchants, totalTransactions int64
	s.db.Model(&models.User{}).Count(&totalUsers)
	s.db.Model(&models.User{}).Where("is_merchant = ?", true).Count(&totalMerchants)
	s.db.Model(&models.Transaction{}).Count(&totalTransactions)

	metrics.TotalUsers = int(totalUsers)
	metrics.TotalMerchants = int(totalMerchants)
	metrics.TotalTransactions = int(totalTransactions)

	return s.db.Save(metrics).Error
}

// UpdateMerchantRankings calculates and stores daily merchant rankings
func (s *MetricsService) UpdateMerchantRankings() error {
	today := time.Now().Format("2006-01-02")

	// Delete existing rankings for today
	s.db.Where("date = ?", today).Delete(&models.MerchantRanking{})

	// Get all merchants with TFI > 0, ordered by TFI
	var merchants []models.User
	err := s.db.Where("is_merchant = ? AND tfi > 0", true).
		Order("tfi DESC").
		Find(&merchants).Error

	if err != nil {
		return fmt.Errorf("failed to get merchants for ranking: %w", err)
	}

	// Create ranking entries
	for i, merchant := range merchants {
		// Get rating count and average for this merchant
		var ratingCount int64
		s.db.Model(&models.Rating{}).Where("merchant_id = ?", merchant.ID).Count(&ratingCount)

		var avgRating struct {
			Avg float64
		}
		s.db.Model(&models.Rating{}).
			Where("merchant_id = ?", merchant.ID).
			Select("AVG((delivery_rating + quality_rating + transparency_rating + environmental_rating) / 4.0) as avg").
			Scan(&avgRating)

		ranking := &models.MerchantRanking{
			MerchantID:    merchant.ID,
			Rank:          i + 1,
			TFI:           merchant.TFI,
			TotalRatings:  int(ratingCount),
			AverageRating: math.Round(avgRating.Avg*100) / 100,
			Date:          today,
			CreatedAt:     time.Now(),
		}

		if err := s.db.Create(ranking).Error; err != nil {
			return fmt.Errorf("failed to create merchant ranking: %w", err)
		}
	}

	return nil
}

// GetMetricsHistory returns historical metrics data for charts
func (s *MetricsService) GetMetricsHistory(days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 30 // Default to last 30 days
	}

	var metrics []models.FairnessMetrics
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	err := s.db.Where("date >= ?", startDate).
		Order("date ASC").
		Find(&metrics).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get metrics history: %w", err)
	}

	// Convert to history format expected by frontend
	pfiHistory := make([]map[string]interface{}, 0)
	tfiHistory := make([]map[string]interface{}, 0)
	cbiHistory := make([]map[string]interface{}, 0)

	// If no historical data exists, generate demo data
	if len(metrics) == 0 {
		return s.generateDemoHistoryData(days), nil
	}

	for _, metric := range metrics {
		// PFI history with percentages
		totalPFI := metric.PFIExcellentCount + metric.PFIGoodCount + metric.PFIAverageCount + metric.PFIPoorCount
		excellentPct := 0.0
		goodPct := 0.0
		averagePct := 0.0
		poorPct := 0.0

		if totalPFI > 0 {
			excellentPct = math.Round((float64(metric.PFIExcellentCount)/float64(totalPFI))*100*100) / 100
			goodPct = math.Round((float64(metric.PFIGoodCount)/float64(totalPFI))*100*100) / 100
			averagePct = math.Round((float64(metric.PFIAverageCount)/float64(totalPFI))*100*100) / 100
			poorPct = math.Round((float64(metric.PFIPoorCount)/float64(totalPFI))*100*100) / 100
		}

		pfiHistory = append(pfiHistory, map[string]interface{}{
			"date":      metric.Date,
			"excellent": excellentPct,
			"good":      goodPct,
			"average":   averagePct,
			"poor":      poorPct,
		})

		// TFI history
		tfiHistory = append(tfiHistory, map[string]interface{}{
			"date":          metric.Date,
			"averageRating": math.Round((metric.AverageTFI/20)*100) / 100, // Convert to 5.0 scale
			"totalRatings":  metric.TotalTFIRatings,
		})

		// CBI history
		cbiHistory = append(cbiHistory, map[string]interface{}{
			"date":  metric.Date,
			"value": math.Round(metric.CBIValue*100) / 100,
		})
	}

	// Get alerts history
	var alerts []models.FairnessAlert
	startTime := time.Now().AddDate(0, 0, -days)
	s.db.Where("created_at >= ?", startTime).
		Order("created_at DESC").
		Find(&alerts)

	alertsHistory := make([]map[string]interface{}, 0)
	for _, alert := range alerts {
		alertsHistory = append(alertsHistory, map[string]interface{}{
			"date":     alert.CreatedAt.Format("2006-01-02"),
			"title":    alert.Title,
			"message":  alert.Description,
			"severity": alert.Severity,
		})
	}

	return map[string]interface{}{
		"pfiHistory":    pfiHistory,
		"tfiHistory":    tfiHistory,
		"cbiHistory":    cbiHistory,
		"alertsHistory": alertsHistory,
	}, nil
}

// CreateAlert creates a fairness alert
func (s *MetricsService) CreateAlert(alertType, severity, title, description string, userID *uuid.UUID) error {
	alert := &models.FairnessAlert{
		Type:        alertType,
		Severity:    severity,
		Title:       title,
		Description: description,
		UserID:      userID,
		IsRead:      false,
		IsResolved:  false,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(alert).Error
}

// GetUnreadAlerts returns unread fairness alerts
func (s *MetricsService) GetUnreadAlerts() ([]models.FairnessAlert, error) {
	var alerts []models.FairnessAlert
	err := s.db.Where("is_read = ?", false).
		Order("created_at DESC").
		Find(&alerts).Error

	return alerts, err
}

// MarkAlertAsRead marks an alert as read
func (s *MetricsService) MarkAlertAsRead(alertID uuid.UUID) error {
	return s.db.Model(&models.FairnessAlert{}).
		Where("id = ?", alertID).
		Update("is_read", true).Error
}

// CheckForAlerts checks for conditions that should trigger alerts
func (s *MetricsService) CheckForAlerts() error {
	// Get the last two days of metrics
	var metrics []models.FairnessMetrics
	err := s.db.Order("date DESC").Limit(2).Find(&metrics).Error
	if err != nil || len(metrics) < 2 {
		return nil // Not enough data
	}

	latest := metrics[0]
	previous := metrics[1]

	// Check for declining fairness trends
	if latest.PFIExcellentCount < previous.PFIExcellentCount && previous.PFIExcellentCount > 0 {
		decline := float64(previous.PFIExcellentCount-latest.PFIExcellentCount) / float64(previous.PFIExcellentCount) * 100
		if decline > 20 { // More than 20% decline
			s.CreateAlert("pfi_decline", "high",
				"Significant PFI Decline Detected",
				fmt.Sprintf("Excellent PFI scores declined by %.1f%% in the last day", decline),
				nil)
		}
	}

	// Check for TFI trends
	if latest.AverageTFI < previous.AverageTFI && previous.AverageTFI > 0 {
		decline := (previous.AverageTFI - latest.AverageTFI) / previous.AverageTFI * 100
		if decline > 10 { // More than 10% decline
			s.CreateAlert("tfi_decline", "medium",
				"TFI Average Declining",
				fmt.Sprintf("Average TFI score declined by %.1f%% in the last day", decline),
				nil)
		}
	}

	// Check for CBI volatility
	change := math.Abs(latest.CBIValue - previous.CBIValue)
	if change > 5 { // More than 5 point change
		severity := "low"
		if change > 10 {
			severity = "high"
		} else if change > 7 {
			severity = "medium"
		}

		direction := "increased"
		if latest.CBIValue < previous.CBIValue {
			direction = "decreased"
		}

		s.CreateAlert("cbi_change", severity,
			"CBI Volatility Alert",
			fmt.Sprintf("Community Basket Index %s by %.1f points", direction, change),
			nil)
	}

	return nil
} // GetComprehensiveMetrics returns all current fairness metrics in one call
func (s *MetricsService) GetComprehensiveMetrics() (map[string]interface{}, error) {
	// Get PFI distribution
	pfiData, err := s.CalculatePFIDistribution()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate PFI distribution: %w", err)
	}

	// Get TFI analysis
	tfiData, err := s.CalculateTFIAnalysis()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate TFI analysis: %w", err)
	}

	// Get top merchants
	topMerchants, err := s.GetTopMerchants(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top merchants: %w", err)
	}

	// Get CBI data
	cbiData, err := s.CalculateCommunityBasketIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate CBI: %w", err)
	}

	// Get recent alerts
	alerts, err := s.GetUnreadAlerts()
	if err != nil {
		alerts = []models.FairnessAlert{} // Don't fail on alerts error
	}

	return map[string]interface{}{
		"pfi_distribution": pfiData["distribution"],
		"tfi_analysis":     tfiData,
		"top_merchants":    topMerchants,
		"cbi":              cbiData,
		"alerts":           alerts,
		"last_updated":     time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// generateDemoHistoryData generates demo historical data when no real data exists
func (s *MetricsService) generateDemoHistoryData(days int) map[string]interface{} {
	pfiHistory := make([]map[string]interface{}, 0)
	tfiHistory := make([]map[string]interface{}, 0)
	cbiHistory := make([]map[string]interface{}, 0)
	alertsHistory := make([]map[string]interface{}, 0)

	// Generate data for the last `days` days
	for i := days; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		// Demo PFI history with some variation
		excellent := 5 + (i % 15)
		good := 25 + (i % 20)
		average := 35 + (i % 15)
		poor := 35 + (i % 25)

		pfiHistory = append(pfiHistory, map[string]interface{}{
			"date":      date,
			"excellent": excellent,
			"good":      good,
			"average":   average,
			"poor":      poor,
		})

		// Demo TFI history
		baseRating := 3.2 + (float64(i%20)-10)/20
		baseRatings := 150 + (i % 100)

		tfiHistory = append(tfiHistory, map[string]interface{}{
			"date":          date,
			"averageRating": math.Round(baseRating*10) / 10,
			"totalRatings":  baseRatings,
		})

		// Demo CBI history with slight variations
		baseValue := 100.0 + (float64(i%20)-10)*2

		cbiHistory = append(cbiHistory, map[string]interface{}{
			"date":  date,
			"value": math.Round(baseValue*10) / 10,
		})

		// Add some demo alerts
		if i%7 == 0 { // Add alert every 7 days
			alertTypes := []map[string]string{
				{"title": "Low Fairness Score", "message": "Merchant fairness score dropped below threshold", "severity": "warning"},
				{"title": "High Transaction Volume", "message": "Unusual transaction volume detected", "severity": "info"},
				{"title": "Rating Anomaly", "message": "Suspicious rating pattern detected", "severity": "critical"},
				{"title": "CBI Fluctuation", "message": "Community Basket Index showing volatility", "severity": "warning"},
			}

			alertType := alertTypes[i%len(alertTypes)]
			alertsHistory = append(alertsHistory, map[string]interface{}{
				"date":     date,
				"title":    alertType["title"],
				"message":  alertType["message"],
				"severity": alertType["severity"],
			})
		}
	}

	return map[string]interface{}{
		"pfiHistory":    pfiHistory,
		"tfiHistory":    tfiHistory,
		"cbiHistory":    cbiHistory,
		"alertsHistory": alertsHistory,
	}
}
