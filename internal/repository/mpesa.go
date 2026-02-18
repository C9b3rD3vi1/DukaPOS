package repository

import (
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"gorm.io/gorm"
)

type MpesaPaymentRepository struct {
	db *gorm.DB
}

func NewMpesaPaymentRepository(db *gorm.DB) *MpesaPaymentRepository {
	return &MpesaPaymentRepository{db: db}
}

func (r *MpesaPaymentRepository) Create(payment *models.MpesaPayment) error {
	return r.db.Create(payment).Error
}

func (r *MpesaPaymentRepository) GetByID(id uint) (*models.MpesaPayment, error) {
	var payment models.MpesaPayment
	err := r.db.First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *MpesaPaymentRepository) GetByCheckoutRequestID(checkoutID string) (*models.MpesaPayment, error) {
	var payment models.MpesaPayment
	err := r.db.Where("checkout_request_id = ?", checkoutID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *MpesaPaymentRepository) GetByMerchantRequestID(merchantID string) (*models.MpesaPayment, error) {
	var payment models.MpesaPayment
	err := r.db.Where("merchant_request_id = ?", merchantID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *MpesaPaymentRepository) GetByPhone(phone string, since time.Time) ([]models.MpesaPayment, error) {
	var payments []models.MpesaPayment
	err := r.db.Where("phone = ? AND created_at > ?", phone, since).Order("created_at DESC").Find(&payments).Error
	return payments, err
}

func (r *MpesaPaymentRepository) GetPendingByShopID(shopID uint, before time.Time) ([]models.MpesaPayment, error) {
	var payments []models.MpesaPayment
	err := r.db.Where("shop_id = ? AND status = ? AND expires_at > ?", shopID, models.MpesaPaymentPending, before).
		Order("created_at ASC").
		Find(&payments).Error
	return payments, err
}

func (r *MpesaPaymentRepository) GetByShopID(shopID uint, limit, offset int) ([]models.MpesaPayment, int64, error) {
	var payments []models.MpesaPayment
	var total int64

	r.db.Model(&models.MpesaPayment{}).Where("shop_id = ?", shopID).Count(&total)
	err := r.db.Where("shop_id = ?", shopID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	return payments, total, err
}

func (r *MpesaPaymentRepository) Update(payment *models.MpesaPayment) error {
	return r.db.Save(payment).Error
}

func (r *MpesaPaymentRepository) UpdateStatus(id uint, status models.MpesaPaymentStatus, receipt, transactionID string) error {
	updates := map[string]interface{}{
		"status":               status,
		"mpesa_receipt":        receipt,
		"mpesa_transaction_id": transactionID,
	}

	if status == models.MpesaPaymentCompleted {
		now := time.Now()
		updates["completed_at"] = now
	}

	return r.db.Model(&models.MpesaPayment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MpesaPaymentRepository) IncrementRetryCount(id uint) error {
	return r.db.Model(&models.MpesaPayment{}).Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

func (r *MpesaPaymentRepository) MarkAsFailed(id uint, reason string) error {
	return r.db.Model(&models.MpesaPayment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         models.MpesaPaymentFailed,
		"failure_reason": reason,
	}).Error
}

func (r *MpesaPaymentRepository) LinkToSale(paymentID, saleID uint) error {
	return r.db.Model(&models.MpesaPayment{}).Where("id = ?", paymentID).Update("sale_id", saleID).Error
}

func (r *MpesaPaymentRepository) Delete(id uint) error {
	return r.db.Delete(&models.MpesaPayment{}, id).Error
}

type MpesaTransactionRepository struct {
	db *gorm.DB
}

func NewMpesaTransactionRepository(db *gorm.DB) *MpesaTransactionRepository {
	return &MpesaTransactionRepository{db: db}
}

func (r *MpesaTransactionRepository) Create(tx *models.MpesaTransaction) error {
	return r.db.Create(tx).Error
}

func (r *MpesaTransactionRepository) GetByTransactionID(transactionID string) (*models.MpesaTransaction, error) {
	var tx models.MpesaTransaction
	err := r.db.Where("transaction_id = ?", transactionID).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *MpesaTransactionRepository) GetByReceiptNumber(receipt string) (*models.MpesaTransaction, error) {
	var tx models.MpesaTransaction
	err := r.db.Where("receipt_number = ?", receipt).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *MpesaTransactionRepository) GetByShopID(shopID uint, limit, offset int) ([]models.MpesaTransaction, int64, error) {
	var transactions []models.MpesaTransaction
	var total int64

	r.db.Model(&models.MpesaTransaction{}).Where("shop_id = ?", shopID).Count(&total)
	err := r.db.Where("shop_id = ?", shopID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}

func (r *MpesaTransactionRepository) Update(tx *models.MpesaTransaction) error {
	return r.db.Save(tx).Error
}

func (r *MpesaTransactionRepository) Delete(id uint) error {
	return r.db.Delete(&models.MpesaTransaction{}, id).Error
}
