package services

import (
	"faircoin/internal/models"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// UserService handles user-related operations
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser creates a new user with a wallet
func (s *UserService) CreateUser(username, email, password, firstName, lastName string) (*models.User, error) {
	user := &models.User{
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		PFI:       10, // Starting PFI score
	}

	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Create user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create wallet for user
	wallet := &models.Wallet{
		UserID:  user.ID,
		Balance: 100.0, // Starting balance for new users
	}

	if err := tx.Create(wallet).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load wallet relationship
	s.db.Preload("Wallet").First(user, user.ID)

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Wallet").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Wallet").First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(userID uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

// GetTopUsersByPFI returns users with highest PFI scores
func (s *UserService) GetTopUsersByPFI(limit int) ([]models.User, error) {
	var users []models.User
	err := s.db.Preload("Wallet").Order("pfi DESC").Limit(limit).Find(&users).Error
	return users, err
}

// GetMerchants returns all verified merchants
func (s *UserService) GetMerchants() ([]models.User, error) {
	var merchants []models.User
	err := s.db.Preload("Wallet").Where("is_merchant = ? AND is_verified = ?", true, true).
		Order("tfi DESC").Find(&merchants).Error
	return merchants, err
}

// RegisterMerchant registers a user as a merchant
func (s *UserService) RegisterMerchant(userID uuid.UUID) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_merchant": true,
			"tfi":         30, // Starting TFI for new merchants
		}).Error
}

// WalletService handles wallet operations
type WalletService struct {
	db *gorm.DB
}

// NewWalletService creates a new wallet service
func NewWalletService(db *gorm.DB) *WalletService {
	return &WalletService{db: db}
}

// GetBalance returns the wallet balance for a user
func (s *WalletService) GetBalance(userID uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet
	err := s.db.Where("user_id = ?", userID).First(&wallet).Error
	return &wallet, err
}

// Transfer transfers FairCoins between users
func (s *WalletService) Transfer(fromUserID, toUserID uuid.UUID, amount float64, description string) (*models.Transaction, error) {
	// Calculate fee (0.1% flat fee)
	fee := amount * 0.001
	if fee < 0.01 {
		fee = 0.01 // Minimum fee
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Check sender balance
	var fromWallet models.Wallet
	if err := tx.Where("user_id = ?", fromUserID).First(&fromWallet).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("sender wallet not found: %w", err)
	}

	if fromWallet.Balance < amount+fee {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient balance")
	}

	// Get receiver wallet
	var toWallet models.Wallet
	if err := tx.Where("user_id = ?", toUserID).First(&toWallet).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("receiver wallet not found: %w", err)
	}

	// Update balances
	fromWallet.Balance -= (amount + fee)
	toWallet.Balance += amount

	if err := tx.Save(&fromWallet).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update sender balance: %w", err)
	}

	if err := tx.Save(&toWallet).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update receiver balance: %w", err)
	}

	// Create transaction record
	transaction := &models.Transaction{
		UserID:      fromUserID,
		ToUserID:    &toUserID,
		Type:        models.TransactionTypeTransfer,
		Amount:      amount,
		Fee:         fee,
		Description: description,
		Status:      "completed",
		CreatedAt:   time.Now(),
	}

	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return transaction, nil
}

// TransactionService handles transaction operations
type TransactionService struct {
	db *gorm.DB
}

// NewTransactionService creates a new transaction service
func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

// GetUserTransactions returns transaction history for a user
func (s *TransactionService) GetUserTransactions(userID uuid.UUID, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := s.db.Preload("User").Preload("ToUser").
		Where("user_id = ? OR to_user_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetTransactionByID returns a specific transaction
func (s *TransactionService) GetTransactionByID(id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	err := s.db.Preload("User").Preload("ToUser").First(&transaction, "id = ?", id).Error
	return &transaction, err
}

// GetCommunityStats returns community statistics
func (s *TransactionService) GetCommunityStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total users
	var userCount int64
	s.db.Model(&models.User{}).Count(&userCount)
	stats["total_users"] = userCount

	// Total merchants
	var merchantCount int64
	s.db.Model(&models.User{}).Where("is_merchant = ?", true).Count(&merchantCount)
	stats["total_merchants"] = merchantCount

	// Total transactions
	var transactionCount int64
	s.db.Model(&models.Transaction{}).Count(&transactionCount)
	stats["total_transactions"] = transactionCount

	// Total circulating supply
	var totalSupply struct {
		Total float64
	}
	s.db.Model(&models.Wallet{}).Select("SUM(balance) as total").Scan(&totalSupply)
	stats["circulating_supply"] = totalSupply.Total

	// Average PFI
	var avgPFI struct {
		Average float64
	}
	s.db.Model(&models.User{}).Select("AVG(pfi) as average").Scan(&avgPFI)
	stats["average_pfi"] = math.Round(avgPFI.Average*100) / 100

	// Transaction volume (last 30 days)
	var volume struct {
		Total float64
	}
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	s.db.Model(&models.Transaction{}).
		Where("created_at > ? AND type = ?", thirtyDaysAgo, models.TransactionTypeTransfer).
		Select("SUM(amount) as total").Scan(&volume)
	stats["transaction_volume_30d"] = volume.Total

	return stats, nil
}
