package api

import (
	"faircoin/internal/config"
	"faircoin/internal/models"
	"faircoin/internal/services"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Handler contains all the services and handles HTTP requests
type Handler struct {
	userService        *services.UserService
	walletService      *services.WalletService
	transactionService *services.TransactionService
	fairnessService    *services.FairnessService
	governanceService  *services.GovernanceService
	monetaryService    *services.MonetaryService
	config             *config.Config
}

// NewHandler creates a new API handler
func NewHandler(
	userService *services.UserService,
	walletService *services.WalletService,
	transactionService *services.TransactionService,
	fairnessService *services.FairnessService,
	governanceService *services.GovernanceService,
	monetaryService *services.MonetaryService,
	cfg *config.Config,
) *Handler {
	return &Handler{
		userService:        userService,
		walletService:      walletService,
		transactionService: transactionService,
		fairnessService:    fairnessService,
		governanceService:  governanceService,
		monetaryService:    monetaryService,
		config:             cfg,
	}
}

// JWT Claims struct
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT tokens
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := bearerToken[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// AdminMiddleware validates admin privileges
func (h *Handler) AdminMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userIDStr, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		user, err := h.userService.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	})
}

// generateToken generates a JWT token for a user
func (h *Handler) generateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWTSecret))
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Username  string `json:"username" binding:"required,min=3,max=50"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=6"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(req.Username, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
		"token":   token,
	})
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := user.CheckPassword(req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// GetProfile returns the user's profile
func (h *Handler) GetProfile(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateProfile updates the user's profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if err := h.userService.UpdateUser(userID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// GetPFI returns the user's PFI breakdown
func (h *Handler) GetPFI(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	breakdown, err := h.fairnessService.GetUserPFIBreakdown(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get PFI breakdown"})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

// AttestUser creates an attestation for another user
func (h *Handler) AttestUser(c *gin.Context) {
	attesterIDStr, _ := c.Get("user_id")
	attesterID, err := uuid.Parse(attesterIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attester ID"})
		return
	}

	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Value       int    `json:"value" binding:"required,min=1,max=10"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	attestation, err := h.fairnessService.CreateAttestation(userID, attesterID, req.Type, req.Value, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Attestation created successfully",
		"attestation": attestation,
	})
}

// GetBalance returns the user's wallet balance
func (h *Handler) GetBalance(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	wallet, err := h.walletService.GetBalance(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// GetTransactionHistory returns the user's transaction history
func (h *Handler) GetTransactionHistory(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 100 {
		limit = 100 // Cap at 100 transactions per request
	}

	transactions, err := h.transactionService.GetUserTransactions(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
	})
}

// SendFairCoins handles FairCoin transfers
func (h *Handler) SendFairCoins(c *gin.Context) {
	fromUserIDStr, _ := c.Get("user_id")
	fromUserID, err := uuid.Parse(fromUserIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sender ID"})
		return
	}

	var req struct {
		ToUsername  string  `json:"to_username" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		Description string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get recipient user
	toUser, err := h.userService.GetUserByUsername(req.ToUsername)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	transaction, err := h.walletService.Transfer(fromUserID, toUser.ID, req.Amount, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transfer successful",
		"transaction": transaction,
	})
}

// GetMerchants returns all verified merchants
func (h *Handler) GetMerchants(c *gin.Context) {
	merchants, err := h.userService.GetMerchants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get merchants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"merchants": merchants,
	})
}

// GetPublicMerchants returns public merchant list (no auth required)
func (h *Handler) GetPublicMerchants(c *gin.Context) {
	merchants, err := h.userService.GetMerchants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get merchants"})
		return
	}

	// Remove sensitive information for public API
	publicMerchants := make([]map[string]interface{}, len(merchants))
	for i, merchant := range merchants {
		publicMerchants[i] = map[string]interface{}{
			"id":         merchant.ID,
			"username":   merchant.Username,
			"first_name": merchant.FirstName,
			"last_name":  merchant.LastName,
			"tfi":        merchant.TFI,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"merchants": publicMerchants,
	})
}

// RegisterMerchant registers the current user as a merchant
func (h *Handler) RegisterMerchant(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.userService.RegisterMerchant(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register as merchant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully registered as merchant",
	})
}

// GetMerchantTFI returns a merchant's TFI breakdown
func (h *Handler) GetMerchantTFI(c *gin.Context) {
	merchantIDStr := c.Param("id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	breakdown, err := h.fairnessService.GetMerchantTFIBreakdown(merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Merchant not found"})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

// RateMerchant creates a rating for a merchant
func (h *Handler) RateMerchant(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	merchantIDStr := c.Param("id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	var req struct {
		DeliveryRating      int    `json:"delivery_rating" binding:"required,min=1,max=10"`
		QualityRating       int    `json:"quality_rating" binding:"required,min=1,max=10"`
		TransparencyRating  int    `json:"transparency_rating" binding:"required,min=1,max=10"`
		EnvironmentalRating int    `json:"environmental_rating" binding:"min=1,max=10"`
		Comments            string `json:"comments"`
		TransactionID       string `json:"transaction_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default environmental rating if not provided
	if req.EnvironmentalRating == 0 {
		req.EnvironmentalRating = 5
	}

	var transactionID *uuid.UUID
	if req.TransactionID != "" {
		txID, err := uuid.Parse(req.TransactionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
			return
		}
		transactionID = &txID
	}

	rating, err := h.fairnessService.CreateRating(
		userID, merchantID, transactionID,
		req.DeliveryRating, req.QualityRating, req.TransparencyRating, req.EnvironmentalRating,
		req.Comments,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rating created successfully",
		"rating":  rating,
	})
}

// GetProposals returns all active governance proposals
func (h *Handler) GetProposals(c *gin.Context) {
	proposals, err := h.governanceService.GetActiveProposals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get proposals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"proposals": proposals,
	})
}

// CreateProposal creates a new governance proposal
func (h *Handler) CreateProposal(c *gin.Context) {
	proposerIDStr, _ := c.Get("user_id")
	proposerID, err := uuid.Parse(proposerIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proposer ID"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required,min=10,max=200"`
		Description string `json:"description" binding:"required,min=50"`
		Type        string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proposalType := models.ProposalType(req.Type)
	proposal, err := h.governanceService.CreateProposal(proposerID, req.Title, req.Description, proposalType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Proposal created successfully",
		"proposal": proposal,
	})
}

// VoteOnProposal handles voting on proposals
func (h *Handler) VoteOnProposal(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	proposalIDStr := c.Param("id")
	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proposal ID"})
		return
	}

	var req struct {
		Vote bool `json:"vote"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.governanceService.VoteOnProposal(userID, proposalID, req.Vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote recorded successfully",
	})
}

// GetCouncilMembers returns the current community council members
func (h *Handler) GetCouncilMembers(c *gin.Context) {
	members, err := h.governanceService.GetCouncilMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get council members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"council_members": members,
	})
}

// GetCommunityStats returns community statistics
func (h *Handler) GetCommunityStats(c *gin.Context) {
	stats, err := h.transactionService.GetCommunityStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get community stats"})
		return
	}

	// Add monetary policy stats
	monthlyStats, err := h.monetaryService.GetCurrentMonthStats()
	if err == nil {
		for key, value := range monthlyStats {
			stats[key] = value
		}
	}

	c.JSON(http.StatusOK, stats)
}

// GetCommunityBasketIndex returns the current CBI
func (h *Handler) GetCommunityBasketIndex(c *gin.Context) {
	cbi, err := h.monetaryService.GetCommunityBasketIndex()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get CBI"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cbi": cbi,
	})
}

// Admin-specific handlers

// GetAdminStats returns comprehensive admin statistics
func (h *Handler) GetAdminStats(c *gin.Context) {
	stats := make(map[string]interface{})

	// Total users
	var totalUsers int64
	if err := h.userService.GetDB().Model(&models.User{}).Count(&totalUsers).Error; err == nil {
		stats["total_users"] = totalUsers
	}

	// Total supply (sum of all wallet balances)
	var totalSupply struct {
		Total float64
	}
	if err := h.walletService.GetDB().Model(&models.Wallet{}).Select("SUM(balance) as total").Scan(&totalSupply).Error; err == nil {
		stats["total_supply"] = totalSupply.Total
	}

	// Daily transactions (last 24 hours)
	var dailyTx int64
	yesterday := time.Now().AddDate(0, 0, -1)
	if err := h.transactionService.GetDB().Model(&models.Transaction{}).
		Where("created_at > ?", yesterday).Count(&dailyTx).Error; err == nil {
		stats["daily_transactions"] = dailyTx
	}

	// Average PFI
	var avgPFI struct {
		Avg float64
	}
	if err := h.userService.GetDB().Model(&models.User{}).Select("AVG(pfi) as avg").Scan(&avgPFI).Error; err == nil {
		stats["avg_pfi"] = avgPFI.Avg
	}

	c.JSON(http.StatusOK, stats)
}

// GetAllUsers returns all users with their details (admin only)
func (h *Handler) GetAllUsers(c *gin.Context) {
	var users []models.User
	if err := h.userService.GetDB().Preload("Wallet").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Transform data for admin view
	var adminUsers []map[string]interface{}
	for _, user := range users {
		adminUser := map[string]interface{}{
			"id":                user.ID,
			"username":          user.Username,
			"email":             user.Email,
			"first_name":        user.FirstName,
			"last_name":         user.LastName,
			"pfi":               user.PFI,
			"tfi":               user.TFI,
			"is_verified":       user.IsVerified,
			"is_merchant":       user.IsMerchant,
			"community_service": user.CommunityService,
			"created_at":        user.CreatedAt,
		}

		// Add wallet balance if available
		if user.Wallet != nil {
			adminUser["balance"] = user.Wallet.Balance
		} else {
			adminUser["balance"] = 0.0
		}

		adminUsers = append(adminUsers, adminUser)
	}

	c.JSON(http.StatusOK, gin.H{"users": adminUsers})
}

// GetAllTransactions returns all transactions (admin only)
func (h *Handler) GetAllTransactions(c *gin.Context) {
	var transactions []models.Transaction
	if err := h.transactionService.GetDB().
		Order("created_at DESC").
		Limit(100). // Limit to last 100 transactions
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	// Transform data for admin view
	var adminTransactions []map[string]interface{}
	for _, tx := range transactions {
		adminTx := map[string]interface{}{
			"id":          tx.ID,
			"type":        tx.Type,
			"amount":      tx.Amount,
			"fee":         tx.Fee,
			"status":      tx.Status,
			"description": tx.Description,
			"created_at":  tx.CreatedAt,
			"user_id":     tx.UserID,
		}

		// Get usernames if needed
		var fromUser models.User
		if err := h.userService.GetDB().First(&fromUser, "id = ?", tx.UserID).Error; err == nil {
			adminTx["from_user"] = fromUser.Username
		}

		if tx.ToUserID != nil {
			var toUser models.User
			if err := h.userService.GetDB().First(&toUser, "id = ?", *tx.ToUserID).Error; err == nil {
				adminTx["to_user"] = toUser.Username
			}
		}

		adminTransactions = append(adminTransactions, adminTx)
	}

	c.JSON(http.StatusOK, gin.H{"transactions": adminTransactions})
}

// GetPFIDistribution returns PFI score distribution
func (h *Handler) GetPFIDistribution(c *gin.Context) {
	var distribution []struct {
		Range string
		Count int64
	}

	// Get counts for different PFI ranges
	var excellent, good, average, poor int64

	h.userService.GetDB().Model(&models.User{}).Where("pfi >= 90").Count(&excellent)
	h.userService.GetDB().Model(&models.User{}).Where("pfi >= 70 AND pfi < 90").Count(&good)
	h.userService.GetDB().Model(&models.User{}).Where("pfi >= 50 AND pfi < 70").Count(&average)
	h.userService.GetDB().Model(&models.User{}).Where("pfi < 50").Count(&poor)

	// Ensure data is in the correct order for the chart
	distribution = append(distribution, struct {
		Range string
		Count int64
	}{"excellent", excellent})
	distribution = append(distribution, struct {
		Range string
		Count int64
	}{"good", good})
	distribution = append(distribution, struct {
		Range string
		Count int64
	}{"average", average})
	distribution = append(distribution, struct {
		Range string
		Count int64
	}{"poor", poor})

	c.JSON(http.StatusOK, gin.H{"distribution": distribution})
}

// GetRecentActivity returns recent system activity
func (h *Handler) GetRecentActivity(c *gin.Context) {
	activities := []map[string]interface{}{}

	// Recent user registrations
	var recentUsers []models.User
	h.userService.GetDB().Order("created_at DESC").Limit(5).Find(&recentUsers)
	for _, user := range recentUsers {
		activities = append(activities, map[string]interface{}{
			"type":    "user_registration",
			"message": "New user " + user.Username + " registered",
			"time":    user.CreatedAt,
			"icon":    "user-plus",
			"color":   "#3498db",
		})
	}

	// Recent transactions
	var recentTx []models.Transaction
	h.transactionService.GetDB().Order("created_at DESC").Limit(5).Find(&recentTx)
	for _, tx := range recentTx {
		activities = append(activities, map[string]interface{}{
			"type":    "transaction",
			"message": "Transaction: " + strconv.FormatFloat(tx.Amount, 'f', 2, 64) + " FC",
			"time":    tx.CreatedAt,
			"icon":    "exchange-alt",
			"color":   "#2ecc71",
		})
	}

	// Recent proposals
	var recentProposals []models.Proposal
	h.governanceService.GetDB().Order("created_at DESC").Limit(3).Find(&recentProposals)
	for _, proposal := range recentProposals {
		activities = append(activities, map[string]interface{}{
			"type":    "proposal",
			"message": "New proposal: " + proposal.Title,
			"time":    proposal.CreatedAt,
			"icon":    "vote-yea",
			"color":   "#f39c12",
		})
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

// UpdateUserStatus allows admin to update user verification status
func (h *Handler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		IsVerified bool `json:"is_verified"`
		PFI        int  `json:"pfi"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Update user
	if err := h.userService.GetDB().Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_verified": req.IsVerified,
			"pfi":         req.PFI,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// GetMonetaryPolicyInfo returns current monetary policy information
func (h *Handler) GetMonetaryPolicyInfo(c *gin.Context) {
	// Get current month's policy
	currentMonth := time.Now().Format("2006-01")
	var policy models.MonetaryPolicy

	err := h.monetaryService.GetDB().Where("month = ?", currentMonth).First(&policy).Error
	if err != nil {
		// Return default values if no policy exists
		c.JSON(http.StatusOK, gin.H{
			"current_month":      currentMonth,
			"base_issuance":      2.0,
			"activity_factor":    1.0,
			"fairness_factor":    1.0,
			"total_issuance":     0.0,
			"circulating_supply": 0.0,
			"average_pfi":        0.0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current_month":      policy.Month,
		"base_issuance":      policy.BaseIssuance,
		"activity_factor":    policy.ActivityFactor,
		"fairness_factor":    policy.FairnessFactor,
		"total_issuance":     policy.TotalIssuance,
		"circulating_supply": policy.CirculatingSupply,
		"average_pfi":        policy.AveragePFI,
	})
}

// GetTransactionVolume returns transaction volume over time
func (h *Handler) GetTransactionVolume(c *gin.Context) {
	var results []struct {
		Month  string
		Volume float64
	}

	// Get transaction volume for the last 6 months
	for i := 5; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

		var monthVolume struct {
			Total float64
		}

		h.transactionService.GetDB().Model(&models.Transaction{}).
			Where("created_at BETWEEN ? AND ? AND type = ?", monthStart, monthEnd, models.TransactionTypeTransfer).
			Select("COALESCE(SUM(amount), 0) as total").Scan(&monthVolume)

		results = append(results, struct {
			Month  string
			Volume float64
		}{
			Month:  date.Format("Jan"),
			Volume: monthVolume.Total,
		})
	}

	c.JSON(http.StatusOK, gin.H{"volume_data": results})
}

// MakeUserAdmin makes the current user an admin (temporary setup method)
func (h *Handler) MakeUserAdmin(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Update user to be admin
	if err := h.userService.GetDB().Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_admin", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make user admin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User is now an admin",
		"user_id": userID,
	})
}
