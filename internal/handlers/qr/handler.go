package handler

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/qr"
	"github.com/gofiber/fiber/v2"
)

type QRHandler struct {
	qrService *qr.QRPaymentService
}

func NewQRHandler(qrSvc *qr.QRPaymentService) *QRHandler {
	return &QRHandler{qrService: qrSvc}
}

func (h *QRHandler) RegisterRoutes(app *fiber.App, protected fiber.Router) {
	qrRoutes := protected.Group("/qr")
	qrRoutes.Post("/generate", h.GenerateDynamicQR)
	qrRoutes.Post("/static", h.GenerateStaticQR)
	qrRoutes.Get("/status/:id", h.GetPaymentStatus)
	qrRoutes.Post("/callback", h.HandleCallback)

	webhook := app.Group("/webhook")
	webhook.Post("/qr", h.HandleCallback)
}

type GenerateQRRequest struct {
	Amount      float64 `json:"amount"`
	Reference   string  `json:"reference"`
	Description string  `json:"description"`
	ProductID   *uint   `json:"product_id"`
	Phone       string  `json:"phone"`
}

func (h *QRHandler) GenerateDynamicQR(c *fiber.Ctx) error {
	if h.qrService == nil {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "QR payment service not configured",
		})
	}

	shopID := c.Locals("shop_id").(uint)

	var req GenerateQRRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Amount must be greater than 0",
		})
	}

	dtoReq := &qr.DynamicQRRequest{
		ShopID:      shopID,
		Amount:      req.Amount,
		Reference:   req.Reference,
		Description: req.Description,
		ProductID:   req.ProductID,
		Phone:       req.Phone,
	}

	resp, err := h.qrService.GenerateDynamicQR(c.Context(), dtoReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"qr_code":    resp.QRCode,
		"amount":     resp.Amount,
		"reference":  resp.Reference,
		"expires_at": resp.ExpiresAt,
		"payment_id": resp.PaymentID,
	})
}

type StaticQRRequest struct {
	ShopID   uint   `json:"shop_id"`
	ShopName string `json:"shop_name"`
}

func (h *QRHandler) GenerateStaticQR(c *fiber.Ctx) error {
	if h.qrService == nil {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "QR payment service not configured",
		})
	}

	shopID := c.Locals("shop_id").(uint)

	var req StaticQRRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	dtoReq := &qr.StaticQRRequest{
		ShopID:   shopID,
		ShopName: req.ShopName,
	}

	resp, err := h.qrService.GenerateStaticQR(dtoReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"qr_code": resp.QRCode,
		"format":  resp.Format,
		"message": resp.Message,
	})
}

func (h *QRHandler) GetPaymentStatus(c *fiber.Ctx) error {
	if h.qrService == nil {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "QR payment service not configured",
		})
	}

	paymentID := c.Params("id")

	payment, err := h.qrService.GetPaymentByCheckoutID(paymentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Payment not found",
		})
	}

	return c.JSON(fiber.Map{
		"id":             payment.ID,
		"amount":         payment.Amount,
		"reference":      payment.Reference,
		"status":         payment.Status,
		"paid_at":        payment.PaidAt,
		"transaction_id": payment.TransactionID,
	})
}

func (h *QRHandler) HandleCallback(c *fiber.Ctx) error {
	if h.qrService == nil {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "QR payment service not configured",
		})
	}

	type CallbackRequest struct {
		CheckoutRequestID string `json:"checkout_request_id"`
		ResultCode        int    `json:"result_code"`
		Amount            string `json:"amount"`
		ReceiptNo         string `json:"receipt_no"`
		PhoneNumber       string `json:"phone_number"`
	}

	var req CallbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid callback data",
		})
	}

	callbackReq := &qr.QRCallbackRequest{
		CheckoutRequestID: req.CheckoutRequestID,
		ResultCode:        req.ResultCode,
		Amount:            req.Amount,
		ReceiptNo:         req.ReceiptNo,
		PhoneNumber:       req.PhoneNumber,
	}

	if err := h.qrService.ProcessCallback(callbackReq); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
	})
}
