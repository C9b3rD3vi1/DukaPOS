package handlers

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PushNotificationHandler struct {
	db *gorm.DB
}

func NewPushNotificationHandler(db *gorm.DB) *PushNotificationHandler {
	return &PushNotificationHandler{db: db}
}

type RegisterDeviceRequest struct {
	DeviceToken string `json:"device_token"`
	Platform    string `json:"platform"` // "ios" or "android"
	ShopID      uint   `json:"shop_id"`
}

type SendPushRequest struct {
	Title  string                 `json:"title"`
	Body   string                 `json:"body"`
	Data   map[string]interface{} `json:"data"`
	ShopID uint                   `json:"shop_id"`
	UserID uint                   `json:"user_id"`
}

func (h *PushNotificationHandler) RegisterRoutes(app fiber.Router) {
	push := app.Group("/notifications")
	push.Post("/register-device", h.RegisterDevice)
	push.Delete("/unregister-device", h.UnregisterDevice)
	push.Post("/send", h.SendPush)
	push.Get("/devices", h.ListDevices)
}

func (h *PushNotificationHandler) RegisterDevice(c *fiber.Ctx) error {
	var req RegisterDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.DeviceToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Device token is required",
		})
	}

	if req.Platform == "" {
		req.Platform = "android" // Default to android
	}

	// Get user ID from context (set by JWT middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Check if device already registered
	var existingDevice models.Device
	result := h.db.Where("device_token = ?", req.DeviceToken).First(&existingDevice)

	if result.Error == nil {
		// Update existing device
		existingDevice.Platform = req.Platform
		existingDevice.UserID = userID.(uint)
		if req.ShopID > 0 {
			existingDevice.ShopID = &req.ShopID
		}
		existingDevice.IsActive = true
		h.db.Save(&existingDevice)
	} else {
		// Create new device
		device := models.Device{
			UserID:      userID.(uint),
			DeviceToken: req.DeviceToken,
			Platform:    req.Platform,
			IsActive:    true,
		}
		if req.ShopID > 0 {
			device.ShopID = &req.ShopID
		}
		h.db.Create(&device)
	}

	return c.JSON(fiber.Map{
		"message": "Device registered successfully",
	})
}

func (h *PushNotificationHandler) UnregisterDevice(c *fiber.Ctx) error {
	var req struct {
		DeviceToken string `json:"device_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result := h.db.Where("device_token = ?", req.DeviceToken).Delete(&models.Device{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to unregister device",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Device unregistered successfully",
	})
}

func (h *PushNotificationHandler) SendPush(c *fiber.Ctx) error {
	var req SendPushRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" || req.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title and body are required",
		})
	}

	// Get devices to send to
	var devices []models.Device
	query := h.db.Where("is_active = ?", true)

	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.ShopID > 0 {
		query = query.Where("shop_id = ?", req.ShopID)
	}

	query.Find(&devices)

	if len(devices) == 0 {
		return c.JSON(fiber.Map{
			"message": "No devices found to send notification",
			"sent":    0,
		})
	}

	// TODO: Implement actual push notification sending
	// This would integrate with FCM for Android and APNs for iOS
	// For now, just log and return success
	sent := 0
	for range devices {
		// In production, send push notification here
		// Example for FCM:
		// fcm.Send(device.DeviceToken, req.Title, req.Body, req.Data)
		sent++
	}

	return c.JSON(fiber.Map{
		"message": "Push notification sent",
		"sent":    sent,
	})
}

func (h *PushNotificationHandler) ListDevices(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var devices []models.Device
	h.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&devices)

	return c.JSON(fiber.Map{
		"devices": devices,
	})
}
