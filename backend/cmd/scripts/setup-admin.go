package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "./faircoin.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// First, add the is_admin column if it doesn't exist
	_, err = db.Exec("ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT false")
	if err != nil {
		log.Printf("Column might already exist: %v", err)
	}

	// Hash password for admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Check if admin user already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check for existing admin: %v", err)
	}

	if count > 0 {
		// Update existing admin user
		_, err = db.Exec(`UPDATE users SET is_admin = true, password_hash = ?, pfi = 100, is_verified = true 
			WHERE username = 'admin'`, string(hashedPassword))
		if err != nil {
			log.Fatalf("Failed to update admin user: %v", err)
		}
		log.Println("Updated existing admin user with admin privileges")
	} else {
		// Create new admin user
		_, err = db.Exec(`INSERT INTO users 
			(id, username, email, password_hash, first_name, last_name, pfi, is_verified, is_admin, is_merchant, tfi, community_service, created_at, updated_at)
			VALUES 
			('550e8400-e29b-41d4-a716-446655440000', 'admin', 'admin@faircoin.com', ?, 'System', 'Administrator', 100, true, true, false, 0, 0, datetime('now'), datetime('now'))`,
			string(hashedPassword))
		if err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
		log.Println("Created new admin user")
	}

	// Also update any existing user to be admin if needed
	log.Println("Setting first user as admin if no admin exists...")
	_, err = db.Exec(`UPDATE users SET is_admin = true WHERE id = (
		SELECT id FROM users WHERE is_admin != true LIMIT 1
	) AND NOT EXISTS (SELECT 1 FROM users WHERE is_admin = true)`)
	if err != nil {
		log.Printf("Warning: Could not set first user as admin: %v", err)
	}

	log.Println("Admin setup completed!")
	log.Println("Login with: username='admin', password='admin123'")
}
