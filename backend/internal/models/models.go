package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// User represents a FairCoin user
type User struct {
	ID               uuid.UUID `json:"id" gorm:"type:varchar(36);primary_key"`
	Username         string    `json:"username" gorm:"unique;not null"`
	Email            string    `json:"email" gorm:"unique;not null"`
	PasswordHash     string    `json:"-" gorm:"not null"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	PFI              int       `json:"pfi" gorm:"default:0"` // Personal Fairness Index (0-100)
	IsVerified       bool      `json:"is_verified" gorm:"default:false"`
	IsMerchant       bool      `json:"is_merchant" gorm:"default:false"`
	IsAdmin          bool      `json:"is_admin" gorm:"default:false"`
	TFI              int       `json:"tfi" gorm:"default:0"`               // Trade Fairness Index (0-100)
	CommunityService int       `json:"community_service" gorm:"default:0"` // Hours of community service
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relations (without circular reference)
	Wallet       *Wallet       `json:"wallet,omitempty" gorm:"foreignkey:UserID"`
	Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignkey:UserID"`
	Attestations []Attestation `json:"attestations,omitempty" gorm:"foreignkey:UserID"`
	Ratings      []Rating      `json:"ratings,omitempty" gorm:"foreignkey:UserID"`
	Proposals    []Proposal    `json:"proposals,omitempty" gorm:"foreignkey:ProposerID"`
	Votes        []Vote        `json:"votes,omitempty" gorm:"foreignkey:UserID"`
}

// Wallet represents a user's FairCoin wallet
type Wallet struct {
	ID        uuid.UUID `json:"id" gorm:"type:varchar(36);primary_key"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null"`
	Balance   float64   `json:"balance" gorm:"default:0"`
	LockedFC  float64   `json:"locked_fc" gorm:"default:0"` // Locked FairCoins (vesting, etc.)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Transaction represents a FairCoin transaction
type Transaction struct {
	ID          uuid.UUID       `json:"id" gorm:"type:varchar(36);primary_key"`
	UserID      uuid.UUID       `json:"user_id" gorm:"type:varchar(36);not null"`
	ToUserID    *uuid.UUID      `json:"to_user_id" gorm:"type:varchar(36)"`
	Type        TransactionType `json:"type" gorm:"not null"`
	Amount      float64         `json:"amount" gorm:"not null"`
	Fee         float64         `json:"fee" gorm:"default:0"`
	Description string          `json:"description"`
	Status      string          `json:"status" gorm:"default:pending"`
	Metadata    string          `json:"metadata" gorm:"type:text"` // JSON metadata
	CreatedAt   time.Time       `json:"created_at"`
}

// TransactionType defines the types of transactions
type TransactionType string

const (
	TransactionTypeTransfer          TransactionType = "transfer"
	TransactionTypeFairnessReward    TransactionType = "fairness_reward"
	TransactionTypeMerchantIncentive TransactionType = "merchant_incentive"
	TransactionTypeMonthlyIssuance   TransactionType = "monthly_issuance"
	TransactionTypeFee               TransactionType = "fee"
	TransactionTypeBurn              TransactionType = "burn"
)

// Attestation represents peer attestations for PFI calculation
type Attestation struct {
	ID          uuid.UUID `json:"id" gorm:"type:varchar(36);primary_key"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null"`     // User being attested
	AttesterID  uuid.UUID `json:"attester_id" gorm:"type:varchar(36);not null"` // User providing attestation
	Type        string    `json:"type" gorm:"not null"`                         // e.g., "community_service", "dispute_resolution", "peer_rating"
	Value       int       `json:"value" gorm:"not null"`                        // 1-10 scale
	Description string    `json:"description"`
	Verified    bool      `json:"verified" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
}

// Rating represents merchant ratings for TFI calculation
type Rating struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:varchar(36);primary_key"`
	UserID              uuid.UUID  `json:"user_id" gorm:"type:varchar(36);not null"`     // Customer
	MerchantID          uuid.UUID  `json:"merchant_id" gorm:"type:varchar(36);not null"` // Merchant being rated
	TransactionID       *uuid.UUID `json:"transaction_id" gorm:"type:varchar(36)"`       // Related transaction
	DeliveryRating      int        `json:"delivery_rating" gorm:"not null"`              // 1-10
	QualityRating       int        `json:"quality_rating" gorm:"not null"`               // 1-10
	TransparencyRating  int        `json:"transparency_rating" gorm:"not null"`          // 1-10
	EnvironmentalRating int        `json:"environmental_rating" gorm:"default:5"`        // 1-10
	Comments            string     `json:"comments"`
	CreatedAt           time.Time  `json:"created_at"`
}

// Proposal represents governance proposals
type Proposal struct {
	ID           uuid.UUID      `json:"id" gorm:"type:varchar(36);primary_key"`
	ProposerID   uuid.UUID      `json:"proposer_id" gorm:"type:varchar(36);not null"`
	Title        string         `json:"title" gorm:"not null"`
	Description  string         `json:"description" gorm:"type:text;not null"`
	Type         ProposalType   `json:"type" gorm:"not null"`
	Status       ProposalStatus `json:"status" gorm:"default:active"`
	VotesFor     int            `json:"votes_for" gorm:"default:0"`
	VotesAgainst int            `json:"votes_against" gorm:"default:0"`
	VotingPower  float64        `json:"voting_power" gorm:"default:0"` // Total voting power participated
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ProposalType defines types of governance proposals
type ProposalType string

const (
	ProposalTypeMonetaryPolicy ProposalType = "monetary_policy"
	ProposalTypeGovernance     ProposalType = "governance"
	ProposalTypeTechnical      ProposalType = "technical"
	ProposalTypeCommunity      ProposalType = "community"
)

// ProposalStatus defines the status of proposals
type ProposalStatus string

const (
	ProposalStatusActive   ProposalStatus = "active"
	ProposalStatusPassed   ProposalStatus = "passed"
	ProposalStatusRejected ProposalStatus = "rejected"
	ProposalStatusExpired  ProposalStatus = "expired"
)

// Vote represents a user's vote on a proposal
type Vote struct {
	ID          uuid.UUID `json:"id" gorm:"type:varchar(36);primary_key"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null"`
	ProposalID  uuid.UUID `json:"proposal_id" gorm:"type:varchar(36);not null"`
	Vote        bool      `json:"vote"`                         // true = for, false = against
	VotingPower float64   `json:"voting_power" gorm:"not null"` // 0.6 * stake_fraction + 0.4 * (PFI/100)
	CreatedAt   time.Time `json:"created_at"`
}

// CommunityBasketIndex represents the community basket index for price stability
type CommunityBasketIndex struct {
	ID           uuid.UUID `json:"id" gorm:"type:varchar(36);primary_key"`
	Value        float64   `json:"value" gorm:"not null"` // Current CBI value
	FoodIndex    float64   `json:"food_index" gorm:"not null"`
	EnergyIndex  float64   `json:"energy_index" gorm:"not null"`
	LaborIndex   float64   `json:"labor_index" gorm:"not null"` // Cost per hour of basic labor
	HousingIndex float64   `json:"housing_index" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
}

// MonetaryPolicy tracks monthly issuance and policy changes
type MonetaryPolicy struct {
	ID                uuid.UUID `json:"id" gorm:"type:varchar(36);primary_key"`
	Month             string    `json:"month" gorm:"not null"` // Format: "2023-10"
	BaseIssuance      float64   `json:"base_issuance" gorm:"not null"`
	ActivityFactor    float64   `json:"activity_factor" gorm:"not null"`
	FairnessFactor    float64   `json:"fairness_factor" gorm:"not null"`
	TotalIssuance     float64   `json:"total_issuance" gorm:"not null"`
	CirculatingSupply float64   `json:"circulating_supply" gorm:"not null"`
	AveragePFI        float64   `json:"average_pfi" gorm:"not null"`
	TotalTransactions int       `json:"total_transactions" gorm:"not null"`
	CreatedAt         time.Time `json:"created_at"`
}

// BeforeCreate sets UUID for models
func (u *User) BeforeCreate(scope *gorm.Scope) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (w *Wallet) BeforeCreate(scope *gorm.Scope) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

func (t *Transaction) BeforeCreate(scope *gorm.Scope) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (a *Attestation) BeforeCreate(scope *gorm.Scope) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (r *Rating) BeforeCreate(scope *gorm.Scope) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (p *Proposal) BeforeCreate(scope *gorm.Scope) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (v *Vote) BeforeCreate(scope *gorm.Scope) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

func (cbi *CommunityBasketIndex) BeforeCreate(scope *gorm.Scope) error {
	if cbi.ID == uuid.Nil {
		cbi.ID = uuid.New()
	}
	return nil
}

func (mp *MonetaryPolicy) BeforeCreate(scope *gorm.Scope) error {
	if mp.ID == uuid.Nil {
		mp.ID = uuid.New()
	}
	return nil
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// CalculateVotingPower calculates a user's voting power based on stake and PFI
func (u *User) CalculateVotingPower(walletBalance, totalSupply float64) float64 {
	// Get user's wallet balance (stake)
	stakePercentage := walletBalance / totalSupply
	pfiPercentage := float64(u.PFI) / 100.0

	// Voting power = 60% stake + 40% PFI
	return 0.6*stakePercentage + 0.4*pfiPercentage
}
