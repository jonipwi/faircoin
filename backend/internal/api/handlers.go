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
