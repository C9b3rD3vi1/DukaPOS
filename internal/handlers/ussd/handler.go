package ussd

import (
	"strconv"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/services/ussd"
	"github.com/gofiber/fiber/v2"
)

// Handler handles USSD HTTP requests
type Handler struct {
	service *ussd.Service
}

// New creates a new USSD handler
func New(service *ussd.Service) *Handler {
	return &Handler{service: service}
}

// USSDRequest represents incoming USSD request
type USSDRequest struct {
	SessionID    string `json:"sessionId"`
	Phone        string `json:"phoneNumber"`
	Text         string `json:"text"`
	NetworkCode  string `json:"networkCode"`
	ServiceCode  string `json:"serviceCode"`
}

// USSDResponse represents USSD response
type USSDResponse struct {
	Response  string `json:"response"`
	SessionID string `json:"sessionId"`
	Action    string `json:"action"` // "continue" or "end"
}

// Handle processes USSD request
// POST /api/v1/ussd
func (h *Handler) Handle(c *fiber.Ctx) error {
	var req USSDRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Generate session ID if not provided
	if req.SessionID == "" {
		req.SessionID = generateSessionID(req.Phone)
	}

	// Process USSD
	response := h.service.Process(req.Phone, req.SessionID, req.Text)

	// Format response
	action := "continue"
	if response.End {
		action = "end"
	}

	return c.JSON(USSDResponse{
		Response:  response.Message,
		SessionID: response.SessionID,
		Action:    action,
	})
}

// HandleAfricaTalking handles USSD from Africa's Talking
// POST /api/v1/ussd/africa
func (h *Handler) HandleAfricaTalking(c *fiber.Ctx) error {
	sessionID := c.FormValue("sessionId")
	phone := c.FormValue("phoneNumber")
	text := c.FormValue("text")

	if sessionID == "" {
		sessionID = generateSessionID(phone)
	}

	response := h.service.Process(phone, sessionID, text)

	action := "continue"
	if response.End {
		action = "end"
	}

	return c.JSON(fiber.Map{
		"response":  response.Message,
		"sessionId": response.SessionID,
		"action":    action,
	})
}

// Callback handles USSD callback
// POST /api/v1/ussd/callback
func (h *Handler) Callback(c *fiber.Ctx) error {
	var req USSDRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func generateSessionID(phone string) string {
	return phone + "_" + strconv.FormatInt(time.Now().Unix(), 10)
}
