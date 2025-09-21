package main

import (
	"faircoin/internal/config"
	"faircoin/internal/database"
	"faircoin/internal/models"
	"log"

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

	// Run migrations to add IsAdmin field
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to migrate User model: %v", err)
	}

	// Create a default admin user if none exists
	var adminCount int64
	db.Model(&models.User{}).Where("is_admin = ?", true).Count(&adminCount)

	if adminCount == 0 {
		log.Println("No admin users found. Creating default admin...")

		// Create default admin user
		adminUser := &models.User{
			Username:   "admin",
			Email:      "admin@faircoin.com",
			FirstName:  "System",
			LastName:   "Administrator",
			PFI:        100,
			IsVerified: true,
			IsAdmin:    true,
		}

		// Set password
		if err := adminUser.SetPassword("admin123"); err != nil {
			log.Fatalf("Failed to set admin password: %v", err)
		}

		// Create user
		if err := db.Create(adminUser).Error; err != nil {
			log.Printf("Warning: Failed to create default admin user: %v", err)
			log.Println("You may need to manually set is_admin=true for an existing user")
		} else {
			log.Printf("Default admin user created: username='admin', password='admin123'")
			log.Println("IMPORTANT: Change the default password after first login!")
		}
	} else {
		log.Printf("Found %d admin users in database", adminCount)
	}

	log.Println("Migration completed successfully")
}
