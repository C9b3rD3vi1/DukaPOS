package qr

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/mpesa"
	"gorm.io/gorm"
)

type QRPaymentService struct {
	db           *gorm.DB
	mpesaSvc     *mpesa.Service
	shopRepo     *repository.ShopRepository
	saleRepo     *repository.SaleRepository
	productRepo  *repository.ProductRepository
	signatureKey []byte
}

type QRPayment struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	ShopID            uint       `gorm:"index" json:"shop_id"`
	CheckoutRequestID string     `gorm:"size:100" json:"checkout_request_id"`
	Amount            float64    `json:"amount"`
	Reference         string     `gorm:"size:100" json:"reference"`
	Phone             string     `gorm:"size:20" json:"phone"`
	Status            string     `gorm:"size:20;default:pending" json:"status"`
	QRCode            string     `gorm:"type:text" json:"qr_code"`
	IsDynamic         bool       `gorm:"default:true" json:"is_dynamic"`
	ExpiresAt         time.Time  `json:"expires_at"`
	PaidAt            *time.Time `json:"paid_at"`
	TransactionID     string     `gorm:"size:100" json:"transaction_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func NewQRPaymentService(
	db *gorm.DB,
	mpesaSvc *mpesa.Service,
	shopRepo *repository.ShopRepository,
	saleRepo *repository.SaleRepository,
	productRepo *repository.ProductRepository,
) *QRPaymentService {
	key := make([]byte, 32)
	copy(key, []byte("DukaPOS-QR-Secret-Key-2024!"))

	svc := &QRPaymentService{
		db:           db,
		mpesaSvc:     mpesaSvc,
		shopRepo:     shopRepo,
		saleRepo:     saleRepo,
		productRepo:  productRepo,
		signatureKey: key,
	}

	db.AutoMigrate(&QRPayment{})
	return svc
}

func (s *QRPaymentService) SavePayment(payment *QRPayment) error {
	return s.db.Create(payment).Error
}

func (s *QRPaymentService) GetPaymentByCheckoutID(checkoutID string) (*QRPayment, error) {
	var payment QRPayment
	err := s.db.Where("checkout_request_id = ?", checkoutID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (s *QRPaymentService) UpdatePaymentStatus(checkoutID string, status string) error {
	return s.db.Model(&QRPayment{}).Where("checkout_request_id = ?", checkoutID).Update("status", status).Error
}

type DynamicQRRequest struct {
	ShopID      uint    `json:"shop_id"`
	Amount      float64 `json:"amount"`
	Reference   string  `json:"reference"`
	Description string  `json:"description"`
	ProductID   *uint   `json:"product_id,omitempty"`
	Phone       string  `json:"phone,omitempty"`
}

type DynamicQRResponse struct {
	QRCode          string    `json:"qr_code"`
	Amount          float64   `json:"amount"`
	Reference       string    `json:"reference"`
	ExpiresAt       time.Time `json:"expires_at"`
	PaymentID       uint      `json:"payment_id"`
	MpesaCheckoutID string    `json:"mpesa_checkout_id,omitempty"`
	Message         string    `json:"message"`
}

type StaticQRRequest struct {
	ShopID   uint   `json:"shop_id"`
	ShopName string `json:"shop_name"`
}

type StaticQRResponse struct {
	QRCode  string `json:"qr_code"`
	Format  string `json:"format"`
	Message string `json:"message"`
}

type QRCallbackRequest struct {
	CheckoutRequestID string `json:"checkout_request_id"`
	MerchantRequestID string `json:"merchant_request_id"`
	ResultCode        int    `json:"result_code"`
	ResultDesc        string `json:"result_desc"`
	Amount            string `json:"amount"`
	ReceiptNo         string `json:"receipt_no"`
	TransactionID     string `json:"transaction_id"`
	PhoneNumber       string `json:"phone_number"`
}

func (s *QRPaymentService) GenerateDynamicQR(ctx context.Context, req *DynamicQRRequest) (*DynamicQRResponse, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if req.Amount > 150000 {
		return nil, fmt.Errorf("amount exceeds maximum (150,000 KES)")
	}

	shop, err := s.shopRepo.GetByID(req.ShopID)
	if err != nil {
		return nil, fmt.Errorf("shop not found")
	}

	reference := req.Reference
	if reference == "" {
		reference = s.generateReference("DQR")
	}

	expiresAt := time.Now().Add(5 * time.Minute)

	qrData := QRCodeData{
		Version:   2,
		ShopID:    strconv.FormatUint(uint64(req.ShopID), 10),
		Amount:    int(req.Amount),
		Reference: reference,
		ProductID: 0,
		Phone:     req.Phone,
		Timestamp: time.Now().Unix(),
		IsDynamic: true,
	}

	signature := s.generateHMACSignature(qrData)
	qrData.Signature = signature

	qrString, err := s.encodeQRData(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR: %w", err)
	}

	paymentID, checkoutID := uint(0), ""

	if s.mpesaSvc != nil && s.mpesaSvc.IsConfigured() {
		phone := req.Phone
		if phone == "" {
			phone = shop.Phone
		}

		validatedPhone, err := s.mpesaSvc.ValidatePhone(phone)
		if err == nil && validatedPhone != "" {
			mpesaReq := &mpesa.PaymentRequest{
				Phone:            validatedPhone,
				Amount:           req.Amount,
				AccountReference: reference,
				Description:      req.Description,
				ShopID:           req.ShopID,
				ProductID:        req.ProductID,
			}

			payment, _, err := s.mpesaSvc.InitiateSTKPush(ctx, mpesaReq)
			if err == nil && payment != nil {
				paymentID = payment.ID
				checkoutID = payment.CheckoutRequestID
			}
		}
	}

	return &DynamicQRResponse{
		QRCode:          qrString,
		Amount:          req.Amount,
		Reference:       reference,
		ExpiresAt:       expiresAt,
		PaymentID:       paymentID,
		MpesaCheckoutID: checkoutID,
		Message:         "QR code generated. Scan with M-Pesa to pay.",
	}, nil
}

func (s *QRPaymentService) GenerateStaticQR(req *StaticQRRequest) (*StaticQRResponse, error) {
	shop, err := s.shopRepo.GetByID(req.ShopID)
	if err != nil {
		return nil, fmt.Errorf("shop not found")
	}

	shopName := req.ShopName
	if shopName == "" {
		shopName = shop.Name
	}

	qrData := QRCodeData{
		Version:   2,
		ShopID:    strconv.FormatUint(uint64(req.ShopID), 10),
		Amount:    0,
		Reference: "STATIC",
		ShopName:  shopName,
		Phone:     shop.Phone,
		Timestamp: time.Now().Unix(),
		IsDynamic: false,
	}

	signature := s.generateHMACSignature(qrData)
	qrData.Signature = signature

	qrString, err := s.encodeQRData(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR: %w", err)
	}

	return &StaticQRResponse{
		QRCode:  qrString,
		Format:  "DukaPOS_Static_v2",
		Message: fmt.Sprintf("Static QR for %s - Pay any amount", shopName),
	}, nil
}

func (s *QRPaymentService) ProcessPayment(ctx context.Context, qrString string, amount float64, phone string) (string, error) {
	qrData, err := s.decodeQRData(qrString)
	if err != nil {
		return "", fmt.Errorf("invalid QR code: %w", err)
	}

	if err := s.verifySignature(qrData); err != nil {
		return "", fmt.Errorf("invalid QR signature: %w", err)
	}

	shopID, err := strconv.ParseUint(qrData.ShopID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid shop ID in QR")
	}

	if qrData.IsDynamic && qrData.Amount > 0 {
		if float64(qrData.Amount) != amount {
			return "", fmt.Errorf("amount mismatch: QR=%d, Paid=%.2f", qrData.Amount, amount)
		}
	}

	shop, err := s.shopRepo.GetByID(uint(shopID))
	if err != nil {
		return "", fmt.Errorf("shop not found")
	}

	reference := qrData.Reference
	if reference == "" {
		reference = s.generateReference("PAY")
	}

	if s.mpesaSvc != nil && s.mpesaSvc.IsConfigured() {
		validatedPhone, err := s.mpesaSvc.ValidatePhone(phone)
		if err == nil {
			mpesaReq := &mpesa.PaymentRequest{
				Phone:            validatedPhone,
				Amount:           amount,
				AccountReference: reference,
				Description:      "QR Payment",
				ShopID:           uint(shopID),
			}

			payment, _, err := s.mpesaSvc.InitiateSTKPush(ctx, mpesaReq)
			if err != nil {
				return "", fmt.Errorf("M-Pesa payment failed: %w", err)
			}

			return fmt.Sprintf("Payment initiated. Check your phone for M-Pesa prompt.\nReference: %s\nCheckoutID: %s",
				reference, payment.CheckoutRequestID), nil
		}
	}

	return fmt.Sprintf("QR Payment received!\nShop: %s\nAmount: KSh %.2f\nReference: %s\n\nThis is a test payment - no real M-Pesa transaction.",
		shop.Name, amount, reference), nil
}

func (s *QRPaymentService) ProcessCallback(req *QRCallbackRequest) error {
	if req.ResultCode != 0 {
		return fmt.Errorf("payment failed: %s", req.ResultDesc)
	}

	return nil
}

func (s *QRPaymentService) VerifyPayment(checkoutID string) (*models.MpesaPayment, error) {
	if s.mpesaSvc == nil {
		return nil, fmt.Errorf("M-Pesa service not configured")
	}

	return s.mpesaSvc.GetPaymentByCheckoutID(checkoutID)
}

func (s *QRPaymentService) generateReference(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s%d", prefix, timestamp)
}

func (s *QRPaymentService) generateHMACSignature(data QRCodeData) string {
	jsonData, _ := json.Marshal(data)
	mac := hmac.New(sha256.New, s.signatureKey)
	mac.Write(jsonData)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (s *QRPaymentService) verifySignature(data *QRCodeData) error {
	if data.IsDynamic && data.Timestamp > 0 {
		expiry := time.Unix(data.Timestamp, 0).Add(5 * time.Minute)
		if time.Now().After(expiry) {
			return fmt.Errorf("QR code expired")
		}
	}

	expectedSig := s.generateHMACSignature(*data)
	if !hmac.Equal([]byte(data.Signature), []byte(expectedSig)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (s *QRPaymentService) encodeQRData(data QRCodeData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(jsonData), nil
}

func (s *QRPaymentService) decodeQRData(qrString string) (*QRCodeData, error) {
	decoded, err := base64.StdEncoding.DecodeString(qrString)
	if err != nil {
		return nil, fmt.Errorf("invalid QR format")
	}

	var data QRCodeData
	if err := json.Unmarshal(decoded, &data); err != nil {
		return nil, fmt.Errorf("invalid QR data")
	}

	return &data, nil
}

type QRCodeData struct {
	Version   int    `json:"v"`
	ShopID    string `json:"sid"`
	Amount    int    `json:"amt"`
	Reference string `json:"ref"`
	ProductID uint   `json:"pid,omitempty"`
	Phone     string `json:"ph,omitempty"`
	ShopName  string `json:"sname,omitempty"`
	Timestamp int64  `json:"ts"`
	Signature string `json:"sig,omitempty"`
	IsDynamic bool   `json:"dyn"`
}

func (s *QRPaymentService) ParseWhatsAppQRCommand(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("Usage: qr pay <amount> or qr generate <amount>")
	}

	command := strings.ToLower(args[0])
	amountStr := args[1]

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return "", fmt.Errorf("invalid amount: %s", amountStr)
	}

	switch command {
	case "pay", "generate":
		return fmt.Sprintf("💳 QR Payment\n\nAmount: KSh %.0f\n\nScan this QR code with M-Pesa to pay.\n\nOr use: pay %s", amount, amountStr), nil
	case "static":
		return "📱 Static QR generated for your shop. Customers can scan and enter any amount.", nil
	default:
		return "", fmt.Errorf("Unknown QR command. Use: qr pay, qr generate, or qr static")
	}
}
