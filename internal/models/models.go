package models

import (
	"time"

	"gorm.io/gorm"
)

// PlanType represents the subscription plan
type PlanType string

const (
	PlanFree     PlanType = "free"
	PlanPro      PlanType = "pro"
	PlanBusiness PlanType = "business"
)

// PaymentMethod represents how payment was made
type PaymentMethod string

const (
	PaymentCash  PaymentMethod = "cash"
	PaymentMpesa PaymentMethod = "mpesa"
	PaymentCard  PaymentMethod = "card"
	PaymentBank  PaymentMethod = "bank"
)

// Account represents an owner account that can own multiple shops
type Account struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	Email               string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash        string         `gorm:"size:255;not null" json:"-"`
	Name                string         `gorm:"size:100;not null" json:"name"`
	Phone               string         `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	IsActive            bool           `gorm:"default:true" json:"is_active"`
	IsVerified          bool           `gorm:"default:false" json:"is_verified"`
	IsAdmin             bool           `gorm:"default:false" json:"is_admin"`
	Plan                PlanType       `gorm:"size:20;default:free" json:"plan"`
	FailedLoginAttempts int            `gorm:"default:0" json:"failed_login_attempts"`
	LockedUntil         *time.Time     `json:"locked_until"`
	LastFailedLogin     *time.Time     `json:"last_failed_login"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Shops []Shop `gorm:"foreignKey:AccountID" json:"shops,omitempty"`
}

// Shop represents a duka/kiosk
type Shop struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	AccountID      uint           `gorm:"index;not null" json:"account_id"`
	Name           string         `gorm:"size:255;not null" json:"name"`
	Phone          string         `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	OwnerName      string         `gorm:"size:100" json:"owner_name"`
	Address        string         `gorm:"size:255" json:"address"`
	Plan           PlanType       `gorm:"size:20;default:free" json:"plan"`
	MpesaShortcode string         `gorm:"size:20" json:"mpesa_shortcode"`
	MpesaPartnerID string         `gorm:"size:50" json:"mpesa_partner_id"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	Email          string         `gorm:"size:100" json:"email"`
	PasswordHash   string         `gorm:"size:255" json:"-"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Account  Account   `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Products []Product `gorm:"foreignKey:ShopID" json:"products,omitempty"`
	Sales    []Sale    `gorm:"foreignKey:ShopID" json:"sales,omitempty"`
	Staff    []Staff   `gorm:"foreignKey:ShopID" json:"staff,omitempty"`
}

// Product represents an item in inventory
type Product struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	ShopID            uint           `gorm:"index;not null" json:"shop_id"`
	Name              string         `gorm:"size:100;not null;index" json:"name"`
	Category          string         `gorm:"size:50" json:"category"`
	Unit              string         `gorm:"size:20;default:pcs" json:"unit"`
	CostPrice         float64        `gorm:"type:decimal(12,2);default:0" json:"cost_price"`
	SellingPrice      float64        `gorm:"type:decimal(12,2);not null" json:"selling_price"`
	Currency          string         `gorm:"size:3;default:KES" json:"currency"`
	AltCurrency       string         `gorm:"size:3" json:"alt_currency"`
	AltPrice          float64        `gorm:"type:decimal(12,2)" json:"alt_price"`
	CurrentStock      int            `gorm:"default:0" json:"current_stock"`
	LowStockThreshold int            `gorm:"default:10" json:"low_stock_threshold"`
	Barcode           string         `gorm:"size:50" json:"barcode"`
	ImageURL          string         `gorm:"size:255" json:"image_url"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Shop  Shop   `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
	Sales []Sale `gorm:"foreignKey:ProductID" json:"sales,omitempty"`
}

// Sale represents a transaction
type Sale struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ShopID        uint           `gorm:"index;not null" json:"shop_id"`
	ProductID     uint           `gorm:"index;not null" json:"product_id"`
	CustomerID    *uint          `gorm:"index" json:"customer_id"`
	Quantity      int            `gorm:"not null" json:"quantity"`
	UnitPrice     float64        `gorm:"type:decimal(12,2);not null" json:"unit_price"`
	TotalAmount   float64        `gorm:"type:decimal(12,2);not null" json:"total_amount"`
	CostAmount    float64        `gorm:"type:decimal(12,2);default:0" json:"cost_amount"`
	Profit        float64        `gorm:"type:decimal(12,2);default:0" json:"profit"`
	PaymentMethod PaymentMethod  `gorm:"size:20;default:cash" json:"payment_method"`
	MpesaReceipt  string         `gorm:"size:50" json:"mpesa_receipt"`
	MpesaPhone    string         `gorm:"size:20" json:"mpesa_phone"`
	StaffID       *uint          `json:"staff_id"`
	Notes         string         `gorm:"size:255" json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Shop     Shop      `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
	Product  Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Staff    *Staff    `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

// DailySummary represents cached daily statistics
type DailySummary struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ShopID            uint      `gorm:"index;not null" json:"shop_id"`
	Date              time.Time `gorm:"type:date;index;not null" json:"date"`
	TotalSales        float64   `gorm:"type:decimal(12,2);default:0" json:"total_sales"`
	TotalTransactions int       `gorm:"default:0" json:"total_transactions"`
	TotalProfit       float64   `gorm:"type:decimal(12,2);default:0" json:"total_profit"`
	TotalCost         float64   `gorm:"type:decimal(12,2);default:0" json:"total_cost"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// Relations
	Shop Shop `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}

// Staff represents staff members
type Staff struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ShopID    uint           `gorm:"index;not null" json:"shop_id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Phone     string         `gorm:"size:20;not null" json:"phone"`
	Role      string         `gorm:"size:50;default:staff" json:"role"`
	Pin       string         `gorm:"size:255" json:"-"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Shop Shop `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}

// Customer represents loyalty customers - see loyalty.go for complete model
// Keeping here for import compatibility - use internal/models/loyalty.go

// Supplier represents suppliers
type Supplier struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ShopID    uint           `gorm:"index;not null" json:"shop_id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Phone     string         `gorm:"size:20" json:"phone"`
	Email     string         `gorm:"size:100" json:"email"`
	Address   string         `gorm:"size:255" json:"address"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Shop Shop `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}

// Order represents supplier orders
type Order struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ShopID      uint           `gorm:"index;not null" json:"shop_id"`
	SupplierID  uint           `gorm:"index;not null" json:"supplier_id"`
	Status      string         `gorm:"size:20;default:pending" json:"status"`
	TotalAmount float64        `gorm:"type:decimal(12,2)" json:"total_amount"`
	Notes       string         `gorm:"size:255" json:"notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Shop     Shop        `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
	Supplier Supplier    `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Items    []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

// OrderItem represents items in an order
type OrderItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	OrderID   uint    `gorm:"index;not null" json:"order_id"`
	ProductID uint    `gorm:"index;not null" json:"product_id"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	UnitCost  float64 `gorm:"type:decimal(12,2)" json:"unit_cost"`
	TotalCost float64 `gorm:"type:decimal(12,2)" json:"total_cost"`

	// Relations
	Order   Order   `gorm:"foreignKey:OrderID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// AuditLog represents system audit logs
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ShopID     uint      `gorm:"index" json:"shop_id"`
	UserType   string    `gorm:"size:20" json:"user_type"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Action     string    `gorm:"size:50;not null" json:"action"`
	EntityType string    `gorm:"size:50" json:"entity_type"`
	EntityID   uint      `json:"entity_id"`
	Details    string    `gorm:"type:text" json:"details"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}

// Webhook represents configured webhooks
type Webhook struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ShopID    uint      `gorm:"index;not null" json:"shop_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	URL       string    `gorm:"size:255;not null" json:"url"`
	Events    string    `gorm:"size:255" json:"events"`
	Secret    string    `gorm:"size:255" json:"secret"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Shop Shop `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}

// APIKey represents API keys for third-party access
type APIKey struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ShopID      uint       `gorm:"index;not null" json:"shop_id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	Key         string     `gorm:"size:50;uniqueIndex;not null" json:"key"`
	SecretHash  string     `gorm:"size:255;not null" json:"-"`
	Permissions string     `gorm:"size:255" json:"permissions"`
	RateLimit   int        `gorm:"default:60" json:"rate_limit"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations
	Shop Shop `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}

// LoyaltyTransaction - see loyalty.go for complete model

// BeforeCreate hook for Shop
func (s *Shop) BeforeCreate(tx *gorm.DB) error {
	if s.Plan == "" {
		s.Plan = PlanFree
	}
	return nil
}

// BeforeCreate hook for Product
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.Unit == "" {
		p.Unit = "pcs"
	}
	if p.LowStockThreshold == 0 {
		p.LowStockThreshold = 10
	}
	return nil
}

// BeforeCreate hook for Sale
func (s *Sale) BeforeCreate(tx *gorm.DB) error {
	if s.PaymentMethod == "" {
		s.PaymentMethod = PaymentCash
	}
	s.Profit = s.TotalAmount - s.CostAmount
	return nil
}
