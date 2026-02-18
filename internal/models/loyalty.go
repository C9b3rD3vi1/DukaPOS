package models

import (
	"time"

	"gorm.io/gorm"
)

type LoyaltyTier string

const (
	TierBronze   LoyaltyTier = "bronze"
	TierSilver   LoyaltyTier = "silver"
	TierGold     LoyaltyTier = "gold"
	TierPlatinum LoyaltyTier = "platinum"
)

type Customer struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ShopID         uint           `gorm:"index;not null" json:"shop_id"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Phone          string         `gorm:"size:20;index" json:"phone"`
	Email          string         `gorm:"size:100" json:"email"`
	Address        string         `gorm:"size:255" json:"address"`
	DateOfBirth    *time.Time     `json:"date_of_birth"`
	LoyaltyPoints  int            `gorm:"default:0" json:"loyalty_points"`
	PointsEarned   int            `gorm:"default:0" json:"points_earned"`
	PointsRedeemed int            `gorm:"default:0" json:"points_redeemed"`
	TotalSpent     float64        `gorm:"default:0" json:"total_spent"`
	Tier           LoyaltyTier    `gorm:"size:20;default:bronze" json:"tier"`
	TotalPurchases int            `gorm:"default:0" json:"total_purchases"`
	LastPurchaseAt *time.Time     `json:"last_purchase_at"`
	ReferralCode   string         `gorm:"size:20;uniqueIndex" json:"referral_code"`
	ReferredBy     *uint          `json:"referred_by"`
	Notes          string         `gorm:"size:500" json:"notes"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	EmailVerified  bool           `gorm:"default:false" json:"email_verified"`
	PhoneVerified  bool           `gorm:"default:false" json:"phone_verified"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Shop         Shop                 `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
	Transactions []LoyaltyTransaction `gorm:"foreignKey:CustomerID" json:"transactions,omitempty"`
}

func (c *Customer) TableName() string {
	return "customers"
}

func (c *Customer) GetTier() LoyaltyTier {
	if c.TotalSpent >= 100000 {
		return TierPlatinum
	} else if c.TotalSpent >= 50000 {
		return TierGold
	} else if c.TotalSpent >= 20000 {
		return TierSilver
	}
	return TierBronze
}

func (c *Customer) ShouldUpgradeTier() bool {
	return c.Tier != c.GetTier()
}

func (c *Customer) UpgradeTier() {
	c.Tier = c.GetTier()
}

type LoyaltyTransactionType string

const (
	LoyaltyEarned   LoyaltyTransactionType = "earned"
	LoyaltyRedeemed LoyaltyTransactionType = "redeemed"
	LoyaltyExpired  LoyaltyTransactionType = "expired"
	LoyaltyBonus    LoyaltyTransactionType = "bonus"
	LoyaltyRefund   LoyaltyTransactionType = "refund"
	LoyaltyAdjust   LoyaltyTransactionType = "adjustment"
)

type LoyaltyTransaction struct {
	ID           uint                   `gorm:"primaryKey" json:"id"`
	CustomerID   uint                   `gorm:"index;not null" json:"customer_id"`
	ShopID       uint                   `gorm:"index;not null" json:"shop_id"`
	SaleID       *uint                  `gorm:"index" json:"sale_id"`
	Type         LoyaltyTransactionType `gorm:"size:20;not null" json:"type"`
	Points       int                    `gorm:"not null" json:"points"`
	PointsBefore int                    `json:"points_before"`
	PointsAfter  int                    `json:"points_after"`
	Amount       float64                `gorm:"type:decimal(12,2)" json:"amount"`
	Description  string                 `gorm:"size:255" json:"description"`
	Reference    string                 `gorm:"size:50;index" json:"reference"`
	ExpiresAt    *time.Time             `json:"expires_at"`
	RedeemedAt   *time.Time             `json:"redeemed_at"`
	CreatedAt    time.Time              `json:"created_at"`

	Customer Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Shop     Shop     `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}

func (t *LoyaltyTransaction) TableName() string {
	return "loyalty_transactions"
}

type LoyaltyReward struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	ShopID        uint        `gorm:"index;not null" json:"shop_id"`
	Name          string      `gorm:"size:100;not null" json:"name"`
	Description   string      `gorm:"size:255" json:"description"`
	PointsCost    int         `gorm:"not null" json:"points_cost"`
	DiscountType  string      `gorm:"size:20" json:"discount_type"`
	DiscountValue float64     `json:"discount_value"`
	MinTier       LoyaltyTier `gorm:"size:20;default:bronze" json:"min_tier"`
	MaxRedeem     int         `gorm:"default:0" json:"max_redeem"`
	RedeemedCount int         `gorm:"default:0" json:"redeemed_count"`
	ValidFrom     *time.Time  `json:"valid_from"`
	ValidUntil    *time.Time  `json:"valid_until"`
	IsActive      bool        `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func (r *LoyaltyReward) TableName() string {
	return "loyalty_rewards"
}

type LoyaltyRedemption struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	CustomerID     uint       `gorm:"index;not null" json:"customer_id"`
	ShopID         uint       `gorm:"index;not null" json:"shop_id"`
	RewardID       uint       `gorm:"index;not null" json:"reward_id"`
	SaleID         *uint      `gorm:"index" json:"sale_id"`
	PointsUsed     int        `gorm:"not null" json:"points_used"`
	DiscountAmount float64    `gorm:"type:decimal(12,2)" json:"discount_amount"`
	Status         string     `gorm:"size:20;default:pending" json:"status"`
	RedeemedAt     *time.Time `json:"redeemed_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (r *LoyaltyRedemption) TableName() string {
	return "loyalty_redemptions"
}

type LoyaltyTierConfig struct {
	Tier        LoyaltyTier `gorm:"size:20;primaryKey" json:"tier"`
	PointsRate  float64     `json:"points_rate"`
	MinSpend    float64     `json:"min_spend"`
	MaxSpend    float64     `json:"max_spend"`
	BonusPoints int         `json:"bonus_points"`
	Perks       string      `gorm:"size:500" json:"perks"`
}

var DefaultTierConfigs = map[LoyaltyTier]LoyaltyTierConfig{
	TierBronze: {
		Tier:        TierBronze,
		PointsRate:  1.0,
		MinSpend:    0,
		MaxSpend:    19999,
		BonusPoints: 0,
		Perks:       "Basic rewards, Birthday bonus points",
	},
	TierSilver: {
		Tier:        TierSilver,
		PointsRate:  1.5,
		MinSpend:    20000,
		MaxSpend:    49999,
		BonusPoints: 100,
		Perks:       "1.5x points, Priority support, 5% birthday discount",
	},
	TierGold: {
		Tier:        TierGold,
		PointsRate:  2.0,
		MinSpend:    50000,
		MaxSpend:    99999,
		BonusPoints: 250,
		Perks:       "2x points, 10% birthday discount, Free delivery",
	},
	TierPlatinum: {
		Tier:        TierPlatinum,
		PointsRate:  3.0,
		MinSpend:    100000,
		MaxSpend:    999999999,
		BonusPoints: 500,
		Perks:       "3x points, 15% birthday discount, Exclusive offers, Personal manager",
	},
}
