package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Port  string
	Debug bool

	// Database configuration
	DBType     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBPath     string

	// JWT configuration
	JWTSecret string

	// Community Basket Index
	CBIUpdateInterval time.Duration
	CBIAPIURL         string

	// Monetary Policy
	BaseMonthlyIssuance  float64
	MaxMonthlyGrowthRate float64
	HoldingCapPercentage float64

	// Fairness System
	MinPFIForProposals       int
	MinTFIForMerchant        int
	AttestationRequiredCount int

	// Security
	BcryptCost        int
	RateLimitRequests int
	RateLimitWindow   int
	AllowedOrigins    string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:  getEnv("PORT", "8080"),
		Debug: getEnvBool("DEBUG", true),

		// Database
		DBType:     getEnv("DB_TYPE", "sqlite"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "faircoin"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "faircoin_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBPath:     getEnv("DB_PATH", "./faircoin.db"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),

		// CBI
		CBIUpdateInterval: getEnvDuration("CBI_UPDATE_INTERVAL", 24*time.Hour),
		CBIAPIURL:         getEnv("CBI_API_URL", "http://localhost:8080/api/v1/cbi"),

		// Monetary Policy
		BaseMonthlyIssuance:  getEnvFloat("BASE_MONTHLY_ISSUANCE", 1000.0),
		MaxMonthlyGrowthRate: getEnvFloat("MAX_MONTHLY_GROWTH_RATE", 0.02),
		HoldingCapPercentage: getEnvFloat("HOLDING_CAP_PERCENTAGE", 0.02),

		// Fairness System
		MinPFIForProposals:       getEnvInt("MIN_PFI_FOR_PROPOSALS", 50),
		MinTFIForMerchant:        getEnvInt("MIN_TFI_FOR_MERCHANT", 30),
		AttestationRequiredCount: getEnvInt("ATTESTATION_REQUIRED_COUNT", 3),

		// Security
		BcryptCost:        getEnvInt("BCRYPT_COST", 12),
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 3600),
		AllowedOrigins:    getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"),
	}
}

// Helper functions to read environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
