package mpesa

import (
	"context"
	"strconv"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/mpesa"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service         *mpesa.Service
	shopRepo        *repository.ShopRepository
	productRepo     *repository.ProductRepository
	saleRepo        *repository.SaleRepository
	paymentRepo     *repository.MpesaPaymentRepository
	transactionRepo *repository.MpesaTransactionRepository
}

func New(
	service *mpesa.Service,
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
	paymentRepo *repository.MpesaPaymentRepository,
	transactionRepo *repository.MpesaTransactionRepository,
) *Handler {
	h := &Handler{
		service:         service,
		shopRepo:        shopRepo,
		productRepo:     productRepo,
		saleRepo:        saleRepo,
		paymentRepo:     paymentRepo,
		transactionRepo: transactionRepo,
	}

	if service != nil {
		service.SetRepositories(paymentRepo, transactionRepo)
		service.SetBusinessRepos(saleRepo, productRepo, shopRepo)
	}

	return h
}

type STKPushRequest struct {
	Phone       string  `json:"phone"`
	Amount      float64 `json:"amount"`
	AccountRef  string  `json:"account_ref"`
	Description string  `json:"description"`
	ProductID   *uint   `json:"product_id"`
}

type STKPushResponse struct {
	Status            string  `json:"status"`
	Message           string  `json:"message"`
	PaymentID         uint    `json:"payment_id"`
	MerchantRequestID string  `json:"merchant_request_id"`
	CheckoutRequestID string  `json:"checkout_request_id"`
	ExpiresIn         int     `json:"expires_in_seconds"`
	Amount            float64 `json:"amount"`
	Phone             string  `json:"phone"`
}

func (h *Handler) STKPush(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{
			"error": "M-Pesa service is not configured. Please set MPESA_CONSUMER_KEY, MPESA_CONSUMER_SECRET, MPESA_SHORTCODE, and MPESA_PASSKEY environment variables.",
		})
	}

	var req STKPushRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if req.Phone == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "phone number is required",
		})
	}

	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "amount must be greater than 0",
		})
	}

	if req.Amount > 150000 {
		return c.Status(400).JSON(fiber.Map{
			"error": "amount exceeds maximum allowed (150,000 KES)",
		})
	}

	shopID := c.Locals("shop_id").(uint)
	if shopID == 0 {
		shop, err := h.shopRepo.GetByPhone(req.Phone)
		if err == nil {
			shopID = shop.ID
		}
	}

	accountRef := req.AccountRef
	if accountRef == "" {
		shop, err := h.shopRepo.GetByID(shopID)
		if err == nil {
			accountRef = shop.Phone
		}
	}

	description := req.Description
	if description == "" {
		description = "DukaPOS Payment"
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	paymentReq := &mpesa.PaymentRequest{
		Phone:            req.Phone,
		Amount:           req.Amount,
		AccountReference: accountRef,
		Description:      description,
		ShopID:           shopID,
		ProductID:        req.ProductID,
	}

	payment, stkResp, err := h.service.InitiateSTKPush(ctx, paymentReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "failed to initiate payment",
			"details": err.Error(),
			"code":    "PAYMENT_INIT_FAILED",
		})
	}

	response := STKPushResponse{
		Status:            "success",
		Message:           "STK push initiated. Please check your phone for the payment prompt.",
		PaymentID:         payment.ID,
		MerchantRequestID: payment.MerchantRequestID,
		CheckoutRequestID: payment.CheckoutRequestID,
		ExpiresIn:         300,
		Amount:            req.Amount,
		Phone:             req.Phone,
	}

	if stkResp != nil && stkResp.CustomerMessage != "" {
		response.Message = stkResp.CustomerMessage
	}

	return c.Status(202).JSON(response)
}

func (h *Handler) GetStatus(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{
			"error": "M-Pesa service is not configured",
		})
	}

	checkoutID := c.Params("checkoutId")
	if checkoutID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "checkout_id is required",
		})
	}

	payment, err := h.service.GetPaymentByCheckoutID(checkoutID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "payment not found",
		})
	}

	return c.JSON(fiber.Map{
		"payment_id":          payment.ID,
		"status":              payment.Status,
		"amount":              payment.Amount,
		"phone":               payment.Phone,
		"checkout_request_id": payment.CheckoutRequestID,
		"merchant_request_id": payment.MerchantRequestID,
		"mpesa_receipt":       payment.MpesaReceipt,
		"failure_reason":      payment.FailureReason,
		"created_at":          payment.CreatedAt,
		"completed_at":        payment.CompletedAt,
		"retry_count":         payment.RetryCount,
	})
}

func (h *Handler) ListPayments(c *fiber.Ctx) error {
	if h.service == nil || h.paymentRepo == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "M-Pesa service not configured",
		})
	}

	shopID := c.Locals("shop_id").(uint)
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	payments, total, err := h.service.GetPaymentsByShop(shopID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch payments",
		})
	}

	return c.JSON(fiber.Map{
		"data":   payments,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) RetryPayment(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{
			"error": "M-Pesa service is not configured",
		})
	}

	paymentID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid payment ID",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	newPayment, err := h.service.RetryPayment(ctx, uint(paymentID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "failed to retry payment",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":              "success",
		"message":             "Payment retry initiated",
		"new_payment_id":      newPayment.ID,
		"checkout_request_id": newPayment.CheckoutRequestID,
	})
}

func (h *Handler) STKCallback(c *fiber.Ctx) error {
	if h.service == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "M-Pesa service not configured",
		})
	}

	body := c.Body()

	payment, err := h.service.ProcessSTKCallback(body)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "failed to process callback",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":         "ok",
		"payment_id":     payment.ID,
		"payment_status": payment.Status,
		"amount":         payment.Amount,
		"mpesa_receipt":  payment.MpesaReceipt,
	})
}

func (h *Handler) C2BCallback(c *fiber.Ctx) error {
	if h.service == nil || h.transactionRepo == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "M-Pesa service not configured",
		})
	}

	var notification mpesa.C2BNotification
	if err := c.BodyParser(&notification); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid callback body",
		})
	}

	tx, err := h.service.HandleC2BNotification(&notification)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to process C2B notification",
		})
	}

	return c.JSON(fiber.Map{
		"status":         "ok",
		"transaction_id": tx.TransactionID,
		"amount":         tx.Amount,
	})
}

func (h *Handler) B2CCallback(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func (h *Handler) BalanceCallback(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func (h *Handler) GetBalance(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Balance query requires B2C API access",
	})
}

type B2CRequest struct {
	Phone   string  `json:"phone"`
	Amount  float64 `json:"amount"`
	Remarks string  `json:"remarks"`
}

func (h *Handler) B2CSend(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(503).JSON(fiber.Map{
			"error": "M-Pesa service not configured",
		})
	}

	var req B2CRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Phone == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "phone is required",
		})
	}
	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "amount must be greater than 0",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "pending",
		"message": "B2C payments require additional Safaricom API access",
		"phone":   req.Phone,
		"amount":  req.Amount,
	})
}

func (h *Handler) GetTransactions(c *fiber.Ctx) error {
	if h.transactionRepo == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "transaction service not configured",
		})
	}

	shopID := c.Locals("shop_id").(uint)
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	transactions, total, err := h.service.GetTransactionsByShop(shopID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch transactions",
		})
	}

	return c.JSON(fiber.Map{
		"data":   transactions,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
