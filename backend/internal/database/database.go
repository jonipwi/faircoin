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

	// For SQLite, handle migration more carefully due to GORM v1 limitations
	if db.Dialect().GetName() == "sqlite3" {
		fmt.Println("Running SQLite database migration...")

		// Try to create tables, but ignore "table exists" errors
		fmt.Println("Ensuring database schema exists...")

		// Try to create each table individually, ignoring "already exists" errors
		tables := []interface{}{
			&models.User{},
			&models.Wallet{},
			&models.Transaction{},
			&models.Attestation{},
			&models.Rating{},
			&models.Proposal{},
			&models.Vote{},
			&models.CommunityBasketIndex{},
			&models.MonetaryPolicy{},
		}

		for _, table := range tables {
			db.AutoMigrate(table) // Ignore errors for existing tables
		}

		// Ensure critical columns exist
		ensureColumn(db, "users", "is_admin", "BOOLEAN DEFAULT false")
		fmt.Println("Database schema update completed")
	} else {
		// For PostgreSQL, AutoMigrate works reliably
		fmt.Println("Running database migration...")
		if err := db.AutoMigrate(
			&models.User{},
			&models.Wallet{},
			&models.Transaction{},
			&models.Attestation{},
			&models.Rating{},
			&models.Proposal{},
			&models.Vote{},
			&models.CommunityBasketIndex{},
			&models.MonetaryPolicy{},
		).Error; err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	fmt.Println("Database migration completed successfully")
	return nil
}

// ensureColumn adds a column to a table if it doesn't exist
func ensureColumn(db *gorm.DB, tableName, columnName, columnType string) {
	// Check if column exists by trying to query it
	var count int
	err := db.Raw(fmt.Sprintf("SELECT COUNT(*) as count FROM pragma_table_info('%s') WHERE name='%s'", tableName, columnName)).Scan(&count).Error
	if err != nil || count == 0 {
		// Column doesn't exist, add it
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
		if err := db.Exec(sql).Error; err != nil {
			fmt.Printf("Warning: Could not add column %s to %s: %v\n", columnName, tableName, err)
		} else {
			fmt.Printf("Added column %s to table %s\n", columnName, tableName)
		}
	}
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
