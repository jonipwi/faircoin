package main

import (
	"faircoin/internal/api"
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/services"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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

	// Auto-migrate database schemas
	if err := database.Migrate(db); err != nil {
		log.Printf("Warning: Database migration encountered issues: %v", err)
		log.Println("Continuing with server startup - this may be expected for existing databases")
	}

	// Initialize services
	userService := services.NewUserService(db)
	walletService := services.NewWalletService(db)
	transactionService := services.NewTransactionService(db)
	fairnessService := services.NewFairnessService(db)
	governanceService := services.NewGovernanceService(db)
	monetaryService := services.NewMonetaryService(db)

	// Start background services
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			// Update PFI/TFI scores
			if err := fairnessService.UpdateAllScores(); err != nil {
				log.Printf("Error updating fairness scores: %v", err)
			}

			// Process monetary policy
			if err := monetaryService.ProcessMonthlyIssuance(); err != nil {
				log.Printf("Error processing monthly issuance: %v", err)
			}
		}
	}()

	// Set up Gin router
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Initialize API routes
	apiHandler := api.NewHandler(
		userService,
		walletService,
		transactionService,
		fairnessService,
		governanceService,
		monetaryService,
		cfg,
	)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication
		auth := v1.Group("/auth")
		{
			auth.POST("/register", apiHandler.Register)
			auth.POST("/login", apiHandler.Login)
			auth.POST("/refresh", apiHandler.RefreshToken)
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(apiHandler.AuthMiddleware())
		{
			users.GET("/profile", apiHandler.GetProfile)
			users.PUT("/profile", apiHandler.UpdateProfile)
			users.GET("/pfi", apiHandler.GetPFI)
			users.POST("/attest", apiHandler.AttestUser)
		}

		// Wallet routes (protected)
		wallet := v1.Group("/wallet")
		wallet.Use(apiHandler.AuthMiddleware())
		{
			wallet.GET("/balance", apiHandler.GetBalance)
			wallet.GET("/history", apiHandler.GetTransactionHistory)
			wallet.POST("/send", apiHandler.SendFairCoins)
		}

		// Merchant routes (protected)
		merchants := v1.Group("/merchants")
		merchants.Use(apiHandler.AuthMiddleware())
		{
			merchants.GET("/", apiHandler.GetMerchants)
			merchants.POST("/register", apiHandler.RegisterMerchant)
			merchants.GET("/:id/tfi", apiHandler.GetMerchantTFI)
			merchants.POST("/:id/rate", apiHandler.RateMerchant)
		}

		// Governance routes (protected)
		governance := v1.Group("/governance")
		governance.Use(apiHandler.AuthMiddleware())
		{
			governance.GET("/proposals", apiHandler.GetProposals)
			governance.POST("/proposals", apiHandler.CreateProposal)
			governance.POST("/proposals/:id/vote", apiHandler.VoteOnProposal)
			governance.GET("/council", apiHandler.GetCouncilMembers)
		}

		// Public routes
		public := v1.Group("/public")
		{
			public.GET("/stats", apiHandler.GetCommunityStats)
			public.GET("/cbi", apiHandler.GetCommunityBasketIndex)
			public.GET("/merchants", apiHandler.GetPublicMerchants)
		}

		// Admin routes (protected with admin privileges)
		admin := v1.Group("/admin")
		admin.Use(apiHandler.AuthMiddleware())
		admin.Use(apiHandler.AdminMiddleware())
		{
			admin.GET("/stats", apiHandler.GetAdminStats)
			admin.GET("/users", apiHandler.GetAllUsers)
			admin.GET("/transactions", apiHandler.GetAllTransactions)
			admin.GET("/pfi-distribution", apiHandler.GetPFIDistribution)
			admin.GET("/activity", apiHandler.GetRecentActivity)
			admin.GET("/transaction-volume", apiHandler.GetTransactionVolume)
			admin.PUT("/users/:id", apiHandler.UpdateUserStatus)
			admin.GET("/monetary-policy", apiHandler.GetMonetaryPolicyInfo)
			admin.POST("/make-admin", apiHandler.MakeUserAdmin) // Temporary endpoint
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// Serve static files for frontend (after API routes to prevent conflicts)
	log.Println("Setting up static file routes...")

	// Check if frontend directory exists
	frontendPath := "../frontend"
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		log.Printf("Warning: %s directory not found", frontendPath)
	} else {
		log.Printf("Frontend directory found at %s", frontendPath)
	}

	// Serve static assets first
	router.Static("/assets", frontendPath+"/assets")

	// Serve frontend pages with explicit routes to avoid conflicts
	router.GET("/", func(c *gin.Context) {
		c.File(frontendPath + "/index.html")
	})

	// Admin dashboard routes (protected)
	adminDashboard := router.Group("/admin-dashboard")
	adminDashboard.Use(apiHandler.AuthMiddleware())
	adminDashboard.Use(apiHandler.AdminMiddleware())
	{
		adminDashboard.GET("/", func(c *gin.Context) {
			c.File(frontendPath + "/admin.html")
		})
	}

	router.GET("/admin.html", apiHandler.AuthMiddleware(), apiHandler.AdminMiddleware(), func(c *gin.Context) {
		c.File(frontendPath + "/admin.html")
	})

	router.GET("/index.html", func(c *gin.Context) {
		c.File(frontendPath + "/index.html")
	})

	log.Println("Static routes configured successfully")

	// Start server
	log.Printf("FairCoin server starting on port %s", cfg.Port)
	log.Printf("Debug mode: %v", cfg.Debug)
	log.Fatal(router.Run(":" + cfg.Port))
}
