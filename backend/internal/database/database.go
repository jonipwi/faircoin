package database

import (
	"faircoin/internal/config"
	"faircoin/internal/models"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "modernc.org/sqlite"
)

// Initialize creates and configures the database connection
func Initialize(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.DBType {
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBPassword, cfg.DBSSLMode)
		db, err = gorm.Open("postgres", dsn)
	case "sqlite":
		db, err = gorm.Open("sqlite", cfg.DBPath)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	// Enable logging in debug mode
	if cfg.Debug {
		db.LogMode(true)
	}

	return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	// Enable UUID extension for PostgreSQL
	if db.Dialect().GetName() == "postgres" {
		db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	}

	// For SQLite, check if tables exist by trying to query one
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		// Table doesn't exist, run migration
		fmt.Println("Running initial database migration...")
		return db.AutoMigrate(
			&models.User{},
			&models.Wallet{},
			&models.Transaction{},
			&models.Attestation{},
			&models.Rating{},
			&models.Proposal{},
			&models.Vote{},
			&models.CommunityBasketIndex{},
			&models.MonetaryPolicy{},
		).Error
	}

	// Tables exist, skip migration
	fmt.Println("Database tables already exist, skipping migration...")
	return nil
} // CreateIndices creates database indices for better performance
func CreateIndices(db *gorm.DB) error {
	// Users indices
	if err := db.Model(&models.User{}).AddIndex("idx_user_username", "username").Error; err != nil {
		return err
	}
	if err := db.Model(&models.User{}).AddIndex("idx_user_email", "email").Error; err != nil {
		return err
	}
	if err := db.Model(&models.User{}).AddIndex("idx_user_pfi", "pfi").Error; err != nil {
		return err
	}

	// Transactions indices
	if err := db.Model(&models.Transaction{}).AddIndex("idx_transaction_user_id", "user_id").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Transaction{}).AddIndex("idx_transaction_created_at", "created_at").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Transaction{}).AddIndex("idx_transaction_type", "type").Error; err != nil {
		return err
	}

	// Attestations indices
	if err := db.Model(&models.Attestation{}).AddIndex("idx_attestation_user_id", "user_id").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Attestation{}).AddIndex("idx_attestation_attester_id", "attester_id").Error; err != nil {
		return err
	}

	// Ratings indices
	if err := db.Model(&models.Rating{}).AddIndex("idx_rating_merchant_id", "merchant_id").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Rating{}).AddIndex("idx_rating_user_id", "user_id").Error; err != nil {
		return err
	}

	// Proposals indices
	if err := db.Model(&models.Proposal{}).AddIndex("idx_proposal_status", "status").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Proposal{}).AddIndex("idx_proposal_created_at", "created_at").Error; err != nil {
		return err
	}

	// Votes indices
	if err := db.Model(&models.Vote{}).AddIndex("idx_vote_proposal_id", "proposal_id").Error; err != nil {
		return err
	}
	if err := db.Model(&models.Vote{}).AddIndex("idx_vote_user_id", "user_id").Error; err != nil {
		return err
	}

	return nil
}
