package models

import (
	"time"

	"gorm.io/gorm"
)

type MpesaPaymentStatus string

const (
	MpesaPaymentPending   MpesaPaymentStatus = "pending"
	MpesaPaymentCompleted MpesaPaymentStatus = "completed"
	MpesaPaymentFailed    MpesaPaymentStatus = "failed"
	MpesaPaymentCancelled MpesaPaymentStatus = "cancelled"
	MpesaPaymentTimeout   MpesaPaymentStatus = "timeout"
)

type MpesaPayment struct {
	ID                 uint               `gorm:"primaryKey" json:"id"`
	ShopID             uint               `gorm:"index;not null" json:"shop_id"`
	ProductID          *uint              `gorm:"index" json:"product_id"`
	Amount             float64            `gorm:"type:decimal(12,2);not null" json:"amount"`
	Phone              string             `gorm:"size:20;index" json:"phone"`
	AccountReference   string             `gorm:"size:50" json:"account_reference"`
	Description        string             `gorm:"size:255" json:"description"`
	MerchantRequestID  string             `gorm:"size:100" json:"merchant_request_id"`
	CheckoutRequestID  string             `gorm:"size:100;uniqueIndex" json:"checkout_request_id"`
	MpesaReceipt       string             `gorm:"size:50" json:"mpesa_receipt"`
	MpesaTransactionID string             `gorm:"size:50" json:"mpesa_transaction_id"`
	Status             MpesaPaymentStatus `gorm:"size:20;default:pending" json:"status"`
	FailureReason      string             `gorm:"size:255" json:"failure_reason"`
	RetryCount         int                `gorm:"default:0" json:"retry_count"`
	SaleID             *uint              `gorm:"index" json:"sale_id"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	CompletedAt        *time.Time         `json:"completed_at"`
	ExpiresAt          time.Time          `json:"expires_at"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"`

	Shop    Shop    `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Sale    *Sale   `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
}

func (m *MpesaPayment) TableName() string {
	return "mpesa_payments"
}

func (m *MpesaPayment) BeforeCreate(tx *gorm.DB) error {
	if m.Status == "" {
		m.Status = MpesaPaymentPending
	}
	if m.ExpiresAt.IsZero() {
		m.ExpiresAt = time.Now().Add(5 * time.Minute)
	}
	return nil
}

type MpesaTransaction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ShopID          uint      `gorm:"index;not null" json:"shop_id"`
	Type            string    `gorm:"size:20;not null" json:"type"` // stk_push, c2b, b2c
	Amount          float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	Phone           string    `gorm:"size:20" json:"phone"`
	TransactionID   string    `gorm:"size:50;uniqueIndex" json:"transaction_id"`
	ReceiptNumber   string    `gorm:"size:50" json:"receipt_number"`
	TransactionTime time.Time `json:"transaction_time"`
	Status          string    `gorm:"size:20" json:"status"`
	CreatedAt       time.Time `json:"created_at"`

	Shop Shop `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
}

func (m *MpesaTransaction) TableName() string {
	return "mpesa_transactions"
}
