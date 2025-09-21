package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"log"
	"os"

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

	log.Println("Checking if is_admin column exists...")

	// Try to add the is_admin column if it doesn't exist
	if err := db.Exec("ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT false").Error; err != nil {
		if err.Error() == "SQL logic error: duplicate column name: is_admin (1)" {
			log.Println("is_admin column already exists")
		} else {
			log.Printf("Error adding is_admin column: %v", err)
		}
	} else {
		log.Println("Successfully added is_admin column")
	}

	// Verify the column exists by trying to query it
	var count int64
	if err := db.Model(&struct{}{}).Table("users").Where("is_admin = ?", false).Count(&count).Error; err != nil {
		log.Printf("Error verifying is_admin column: %v", err)
		os.Exit(1)
	}

	log.Printf("is_admin column verified - found %d non-admin users", count)
	log.Println("Database column fix completed successfully!")
}
