package main

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID               string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Username         string `gorm:"uniqueIndex;not null" json:"username"`
	Email            string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash     string `gorm:"not null" json:"-"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	PFI              int    `gorm:"default:10" json:"pfi"`
	IsVerified       bool   `gorm:"default:false" json:"is_verified"`
	IsMerchant       bool   `gorm:"default:false" json:"is_merchant"`
	IsAdmin          bool   `gorm:"default:false" json:"is_admin"`
	TFI              int    `gorm:"default:0" json:"tfi"`
	CommunityService int    `gorm:"default:0" json:"community_service"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func main() {
	// Open database connection
	db, err := gorm.Open(sqlite.Open("faircoin.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Query all users
	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		log.Fatal("Failed to query users:", result.Error)
	}

	fmt.Printf("Found %d users:\n", len(users))
	fmt.Printf("%-15s %-15s %-5s %-5s %-10s %-10s %-8s %-15s\n",
		"Username", "Email", "PFI", "TFI", "Verified", "Merchant", "Admin", "CommService")
	fmt.Println(strings.Repeat("-", 90))

	for _, user := range users {
		fmt.Printf("%-15s %-15s %-5d %-5d %-10t %-10t %-8t %-15d\n",
			user.Username, user.Email, user.PFI, user.TFI,
			user.IsVerified, user.IsMerchant, user.IsAdmin, user.CommunityService)
	}
}
