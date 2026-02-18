package mpesa

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
)

var (
	ErrMpesaNotConfigured = errors.New("M-Pesa is not configured")
	ErrInvalidPhone       = errors.New("invalid phone number")
	ErrPaymentExpired     = errors.New("payment request expired")
	ErrPaymentFailed      = errors.New("payment failed")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidCredentials = errors.New("invalid M-Pesa credentials")
	ErrRateLimited        = errors.New("M-Pesa API rate limited")
	ErrNetworkError       = errors.New("network error connecting to M-Pesa")
)

const (
	MaxRetries          = 3
	TokenCacheDuration  = 50 * time.Minute
	PaymentTimeout      = 5 * time.Minute
	STKPushEndpoint     = "mpesa/stkpush/v1/processrequest"
	STKQueryEndpoint    = "mpesa/stkpushquery/v1/query"
	OAuthEndpoint       = "oauth/v1/generate"
	RegisterURLEndpoint = "mpesa/c2b/v1/registerurl"
	C2BEndpoint         = "mpesa/c2b/v1/simulate"
	B2CEndpoint         = "mpesa/b2c/v1/paymentrequest"
)

type Config struct {
	ConsumerKey        string
	ConsumerSecret     string
	Shortcode          string
	Passkey            string
	CallbackURL        string
	Environment        string
	InitiatorName      string
	SecurityCredential string
}

type Service struct {
	config          *Config
	httpClient      *http.Client
	authToken       string
	tokenExpiry     time.Time
	tokenMutex      sync.RWMutex
	paymentRepo     *repository.MpesaPaymentRepository
	transactionRepo *repository.MpesaTransactionRepository
	saleRepo        *repository.SaleRepository
	productRepo     *repository.ProductRepository
	shopRepo        *repository.ShopRepository
	callbackURL     string
	isConfigured    bool
	environment     string
}

type PaymentRequest struct {
	Phone            string
	Amount           float64
	AccountReference string
	Description      string
	ShopID           uint
	ProductID        *uint
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

type CallbackData struct {
	Amount                   string `json:"Amount"`
	ReceiptNo                string `json:"ReceiptNo"`
	TransactionID            string `json:"TransactionID"`
	PhoneNumber              string `json:"PhoneNumber"`
	TransactionDate          string `json:"TransactionDate"`
	TransactionTime          string `json:"TransactionTime"`
	ConversationID           string `json:"ConversationID"`
	OriginatorConversationID string `json:"OriginatorConversationID"`
}

type STKCallback struct {
	MerchantRequestID string           `json:"MerchantRequestID"`
	CheckoutRequestID string           `json:"CheckoutRequestID"`
	ResultCode        int              `json:"ResultCode"`
	ResultDesc        string           `json:"ResultDesc"`
	CallbackMetadata  CallbackMetadata `json:"CallbackMetadata"`
}

type CallbackMetadata struct {
	Item []CallbackItem `json:"Item"`
}

type CallbackItem struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type C2BNotification struct {
	TransactionType     string `json:"TransactionType"`
	TransactionID       string `json:"TransactionID"`
	TransactionTime     string `json:"TransactionTime"`
	Amount              string `json:"Amount"`
	BusinessShortCode   string `json:"BusinessShortCode"`
	BillReferenceNumber string `json:"BillRefNumber"`
	InvoiceNumber       string `json:"InvoiceNumber"`
	ExternalIdentifier  string `json:"ExternalIdentifier"`
	PhoneNumber         string `json:"PhoneNumber"`
	AccountReference    string `json:"AccountReference"`
	Name                string `json:"Name"`
}

func New(config *Config, paymentRepo *repository.MpesaPaymentRepository, transactionRepo *repository.MpesaTransactionRepository) *Service {
	if config == nil {
		config = &Config{Environment: "sandbox"}
	}

	svc := &Service{
		config:      config,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		callbackURL: config.CallbackURL,
		environment: config.Environment,
	}

	if config.ConsumerKey != "" && config.ConsumerSecret != "" && config.Shortcode != "" {
		svc.isConfigured = true
	}

	svc.paymentRepo = paymentRepo
	svc.transactionRepo = transactionRepo

	return svc
}

func (s *Service) SetRepositories(paymentRepo *repository.MpesaPaymentRepository, transactionRepo *repository.MpesaTransactionRepository) {
	s.paymentRepo = paymentRepo
	s.transactionRepo = transactionRepo
}

func (s *Service) SetBusinessRepos(saleRepo *repository.SaleRepository, productRepo *repository.ProductRepository, shopRepo *repository.ShopRepository) {
	s.saleRepo = saleRepo
	s.productRepo = productRepo
	s.shopRepo = shopRepo
}

func (s *Service) IsConfigured() bool {
	return s.isConfigured
}

func (s *Service) getBaseURL() string {
	if s.environment == "live" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

func (s *Service) getAuthURL() string {
	return fmt.Sprintf("%s/%s", s.getBaseURL(), OAuthEndpoint)
}

func (s *Service) getSTKPushURL() string {
	return fmt.Sprintf("%s/%s", s.getBaseURL(), STKPushEndpoint)
}

func (s *Service) getSTKQueryURL() string {
	return fmt.Sprintf("%s/%s", s.getBaseURL(), STKQueryEndpoint)
}

func (s *Service) getToken() (string, error) {
	s.tokenMutex.RLock()
	if s.authToken != "" && time.Now().Before(s.tokenExpiry) {
		defer s.tokenMutex.RUnlock()
		return s.authToken, nil
	}
	s.tokenMutex.RUnlock()

	return s.getTokenFresh()
}

func (s *Service) getTokenFresh() (string, error) {
	s.tokenMutex.Lock()
	defer s.tokenMutex.Unlock()

	if s.authToken != "" && time.Now().Before(s.tokenExpiry) {
		return s.authToken, nil
	}

	credentials := fmt.Sprintf("%s.config.ConsumerKey, s.config.ConsumerSecret:%s", s)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))

	req, err := http.NewRequest("GET", s.getAuthURL(), nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNetworkError, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return "", ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("M-Pesa auth failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if result.AccessToken == "" {
		return "", ErrInvalidCredentials
	}

	s.authToken = result.AccessToken
	expiresIn := 3600
	if result.ExpiresIn != "" {
		if e, err := strconv.Atoi(result.ExpiresIn); err == nil {
			expiresIn = e
		}
	}
	s.tokenExpiry = time.Now().Add(time.Duration(expiresIn-300) * time.Second)

	return s.authToken, nil
}

func (s *Service) ValidatePhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	phone = strings.TrimPrefix(phone, "+")

	if strings.HasPrefix(phone, "254") && len(phone) == 12 {
		return phone, nil
	}

	if strings.HasPrefix(phone, "0") && len(phone) == 10 {
		return "254" + phone[1:], nil
	}

	if strings.HasPrefix(phone, "7") && len(phone) == 9 {
		return "254" + phone, nil
	}

	if strings.HasPrefix(phone, "1") && len(phone) == 9 {
		return "254" + phone, nil
	}

	if matched, _ := regexp.MatchString(`^254[0-9]{9}$`, phone); matched {
		return phone, nil
	}

	return "", ErrInvalidPhone
}

func (s *Service) GeneratePassword(timestamp string) string {
	data := fmt.Sprintf("%s%s%s", s.config.Shortcode, s.config.Passkey, timestamp)
	hash := md5.Sum([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (s *Service) InitiateSTKPush(ctx context.Context, req *PaymentRequest) (*models.MpesaPayment, *STKPushResponse, error) {
	if !s.isConfigured {
		return nil, nil, ErrMpesaNotConfigured
	}

	validatedPhone, err := s.ValidatePhone(req.Phone)
	if err != nil {
		return nil, nil, err
	}

	if req.Amount <= 0 {
		return nil, nil, errors.New("amount must be greater than 0")
	}

	if req.Amount > 150000 {
		return nil, nil, errors.New("amount exceeds maximum allowed (150,000 KES)")
	}

	payment := &models.MpesaPayment{
		ShopID:           req.ShopID,
		ProductID:        req.ProductID,
		Amount:           req.Amount,
		Phone:            validatedPhone,
		AccountReference: req.AccountReference,
		Description:      req.Description,
		Status:           models.MpesaPaymentPending,
		ExpiresAt:        time.Now().Add(PaymentTimeout),
	}

	token, err := s.getToken()
	if err != nil {
		payment.Status = models.MpesaPaymentFailed
		payment.FailureReason = fmt.Sprintf("Auth failed: %v", err)
		if s.paymentRepo != nil {
			_ = s.paymentRepo.Create(payment)
		}
		return payment, nil, err
	}

	timestamp := time.Now().Format("20060102150405")
	password := s.GeneratePassword(timestamp)

	stkReq := map[string]interface{}{
		"BusinessShortCode": s.config.Shortcode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            int(req.Amount),
		"PartyA":            validatedPhone,
		"PartyB":            s.config.Shortcode,
		"PhoneNumber":       validatedPhone,
		"CallBackURL":       s.callbackURL,
		"AccountReference":  req.AccountReference,
		"TransactionDesc":   req.Description,
	}

	body, err := json.Marshal(stkReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.getSTKPushURL(), bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		payment.Status = models.MpesaPaymentFailed
		payment.FailureReason = fmt.Sprintf("Network error: %v", err)
		if s.paymentRepo != nil {
			_ = s.paymentRepo.Create(payment)
		}
		return payment, nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result STKPushResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		payment.Status = models.MpesaPaymentFailed
		payment.FailureReason = fmt.Sprintf("Invalid response: %v", err)
		if s.paymentRepo != nil {
			_ = s.paymentRepo.Create(payment)
		}
		return payment, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	payment.MerchantRequestID = result.MerchantRequestID
	payment.CheckoutRequestID = result.CheckoutRequestID

	if result.ResponseCode != "0" {
		payment.Status = models.MpesaPaymentFailed
		payment.FailureReason = result.ResponseDescription

		if result.ResponseCode == "1" {
			payment.FailureReason = "M-Pesa is currently unavailable"
		} else if result.ResponseCode == "2" {
			payment.FailureReason = "Invalid M-Pesa credentials"
		} else if result.ResponseCode == "3" {
			payment.FailureReason = "Invalid shortcode"
		} else if result.ResponseCode == "4" {
			payment.FailureReason = "Invalid transaction type"
		} else if result.ResponseCode == "5" {
			payment.FailureReason = "Invalid amount"
		} else if result.ResponseCode == "6" {
			payment.FailureReason = "Invalid party"
		} else if result.ResponseCode == "17" {
			payment.FailureReason = "Invalid SMS sender"
		}

		if s.paymentRepo != nil {
			_ = s.paymentRepo.Create(payment)
		}

		return payment, &result, fmt.Errorf("STK push failed: %s", result.ResponseDescription)
	}

	if s.paymentRepo != nil {
		_ = s.paymentRepo.Create(payment)
	}

	return payment, &result, nil
}

func (s *Service) QuerySTKStatus(ctx context.Context, checkoutID string) (*STKPushResponse, error) {
	if !s.isConfigured {
		return nil, ErrMpesaNotConfigured
	}

	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102150405")
	password := s.GeneratePassword(timestamp)

	queryReq := map[string]interface{}{
		"BusinessShortCode": s.config.Shortcode,
		"Password":          password,
		"Timestamp":         timestamp,
		"CheckoutRequestID": checkoutID,
	}

	body, _ := json.Marshal(queryReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.getSTKQueryURL(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}
	defer resp.Body.Close()

	var result STKPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (s *Service) ProcessSTKCallback(callbackBody []byte) (*models.MpesaPayment, error) {
	var callback struct {
		Body struct {
			STKCallback STKCallback `json:"stkCallback"`
		} `json:"body"`
	}

	if err := json.Unmarshal(callbackBody, &callback); err != nil {
		return nil, fmt.Errorf("failed to parse callback: %w", err)
	}

	stkCallback := callback.Body.STKCallback

	payment, err := s.paymentRepo.GetByCheckoutRequestID(stkCallback.CheckoutRequestID)
	if err != nil {
		return nil, fmt.Errorf("payment not found for checkout: %s", stkCallback.CheckoutRequestID)
	}

	if stkCallback.ResultCode == 0 {
		receipt := ""
		transactionID := ""

		for _, item := range stkCallback.CallbackMetadata.Item {
			switch item.Name {
			case "Amount":
			case "MpesaReceiptNumber":
				receipt = item.Value
			case "TransactionID":
				transactionID = item.Value
			}
		}

		payment.MpesaReceipt = receipt
		payment.MpesaTransactionID = transactionID
		payment.Status = models.MpesaPaymentCompleted

		now := time.Now()
		payment.CompletedAt = &now

		if err := s.paymentRepo.Update(payment); err != nil {
			return nil, fmt.Errorf("failed to update payment: %w", err)
		}

		if s.saleRepo != nil && s.productRepo != nil && payment.ProductID != nil {
			s.processSuccessfulPayment(payment)
		}

		_ = s.transactionRepo.Create(&models.MpesaTransaction{
			ShopID:          payment.ShopID,
			Type:            "stk_push",
			Amount:          payment.Amount,
			Phone:           payment.Phone,
			TransactionID:   transactionID,
			ReceiptNumber:   receipt,
			TransactionTime: time.Now(),
			Status:          "completed",
		})

	} else {
		payment.Status = models.MpesaPaymentFailed
		payment.FailureReason = stkCallback.ResultDesc
		_ = s.paymentRepo.Update(payment)
	}

	return payment, nil
}

func (s *Service) processSuccessfulPayment(payment *models.MpesaPayment) {
	product, err := s.productRepo.GetByID(*payment.ProductID)
	if err != nil {
		return
	}

	if product.CurrentStock < 1 {
		return
	}

	qty := 1
	if payment.Amount >= product.SellingPrice*2 {
		qty = int(payment.Amount / product.SellingPrice)
		if qty > product.CurrentStock {
			qty = product.CurrentStock
		}
	}

	totalAmount := product.SellingPrice * float64(qty)
	costAmount := product.CostPrice * float64(qty)
	profit := totalAmount - costAmount

	sale := &models.Sale{
		ShopID:        payment.ShopID,
		ProductID:     product.ID,
		Quantity:      qty,
		UnitPrice:     product.SellingPrice,
		TotalAmount:   totalAmount,
		CostAmount:    costAmount,
		Profit:        profit,
		PaymentMethod: models.PaymentMpesa,
		MpesaReceipt:  payment.MpesaReceipt,
		MpesaPhone:    payment.Phone,
		Notes:         fmt.Sprintf("M-Pesa Payment: %s", payment.MpesaReceipt),
	}

	if err := s.saleRepo.Create(sale); err != nil {
		return
	}

	_ = s.productRepo.UpdateStock(product.ID, -qty)
	_ = s.paymentRepo.LinkToSale(payment.ID, sale.ID)
}

func (s *Service) HandleC2BNotification(notification *C2BNotification) (*models.MpesaTransaction, error) {
	amount, _ := strconv.ParseFloat(notification.Amount, 64)
	phone := strings.TrimPrefix(notification.PhoneNumber, "+")

	existing, _ := s.transactionRepo.GetByReceiptNumber(notification.BillReferenceNumber)
	if existing != nil {
		return existing, nil
	}

	tx := &models.MpesaTransaction{
		TransactionID:   notification.TransactionID,
		ReceiptNumber:   notification.BillReferenceNumber,
		Amount:          amount,
		Phone:           phone,
		Type:            "c2b",
		TransactionTime: time.Now(),
		Status:          "completed",
	}

	if s.transactionRepo != nil {
		if err := s.transactionRepo.Create(tx); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (s *Service) GetPaymentByID(id uint) (*models.MpesaPayment, error) {
	if s.paymentRepo == nil {
		return nil, errors.New("payment repository not configured")
	}
	return s.paymentRepo.GetByID(id)
}

func (s *Service) GetPaymentByCheckoutID(checkoutID string) (*models.MpesaPayment, error) {
	if s.paymentRepo == nil {
		return nil, errors.New("payment repository not configured")
	}
	return s.paymentRepo.GetByCheckoutRequestID(checkoutID)
}

func (s *Service) GetPaymentsByShop(shopID uint, limit, offset int) ([]models.MpesaPayment, int64, error) {
	if s.paymentRepo == nil {
		return nil, 0, errors.New("payment repository not configured")
	}
	return s.paymentRepo.GetByShopID(shopID, limit, offset)
}

func (s *Service) GetTransactionsByShop(shopID uint, limit, offset int) ([]models.MpesaTransaction, int64, error) {
	if s.transactionRepo == nil {
		return nil, 0, errors.New("transaction repository not configured")
	}
	return s.transactionRepo.GetByShopID(shopID, limit, offset)
}

func (s *Service) RetryPayment(ctx context.Context, paymentID uint) (*models.MpesaPayment, error) {
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, err
	}

	if payment.Status != models.MpesaPaymentPending {
		return nil, errors.New("payment is not in pending state")
	}

	if time.Now().After(payment.ExpiresAt) {
		_ = s.paymentRepo.MarkAsFailed(paymentID, "Payment request expired")
		return nil, ErrPaymentExpired
	}

	if payment.RetryCount >= MaxRetries {
		_ = s.paymentRepo.MarkAsFailed(paymentID, "Maximum retries exceeded")
		return nil, errors.New("maximum retry attempts exceeded")
	}

	req := &PaymentRequest{
		Phone:            payment.Phone,
		Amount:           payment.Amount,
		AccountReference: payment.AccountReference,
		Description:      payment.Description,
		ShopID:           payment.ShopID,
	}

	newPayment, _, err := s.InitiateSTKPush(ctx, req)
	if err != nil {
		_ = s.paymentRepo.IncrementRetryCount(paymentID)
		return nil, err
	}

	return newPayment, nil
}

func (s *Service) ProcessExpiredPayments() error {
	now := time.Now()
	shops, _, err := s.shopRepo.List(1000, 0)
	if err != nil {
		return err
	}

	for _, shop := range shops {
		payments, err := s.paymentRepo.GetPendingByShopID(shop.ID, now)
		if err != nil {
			continue
		}

		for _, payment := range payments {
			if now.After(payment.ExpiresAt) {
				_ = s.paymentRepo.MarkAsFailed(payment.ID, "Payment request expired")
			}
		}
	}

	return nil
}

func ParseCallback(data []byte) (*CallbackData, error) {
	var callback struct {
		Body struct {
			CallbackMetadata struct {
				Item []struct {
					Name  string `json:"Name"`
					Value string `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"Body"`
	}

	if err := json.Unmarshal(data, &callback); err != nil {
		return nil, err
	}

	result := &CallbackData{}
	for _, item := range callback.Body.CallbackMetadata.Item {
		switch item.Name {
		case "Amount":
			result.Amount = item.Value
		case "MpesaReceiptNumber", "BillRefNumber":
			result.ReceiptNo = item.Value
		case "TransactionID":
			result.TransactionID = item.Value
		case "PhoneNumber":
			result.PhoneNumber = item.Value
		case "TransactionDate":
			result.TransactionDate = item.Value
		}
	}

	return result, nil
}

func GenerateTransactionID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("TXN%d%d", time.Now().Unix(), rand.Intn(10000))
}
