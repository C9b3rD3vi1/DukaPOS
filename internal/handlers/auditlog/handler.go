package handlers

import (
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/gofiber/fiber/v2"
)

type AuditLogHandler struct {
	auditRepo *repository.AuditLogRepository
}

func NewAuditLogHandler(auditRepo *repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{
		auditRepo: auditRepo,
	}
}

func (h *AuditLogHandler) RegisterRoutes(app fiber.Router) {
	audit := app.Group("/audit-logs")
	audit.Get("/", h.GetLogs)
	audit.Get("/:id", h.GetLog)
	audit.Get("/user/:userID", h.GetByUser)
	audit.Get("/entity/:type/:id", h.GetByEntity)
	audit.Get("/action/:action", h.GetByAction)
	audit.Get("/stats/summary", h.GetStatsSummary)
}

func (h *AuditLogHandler) GetLogs(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}

	logs, err := h.auditRepo.GetByShopID(shopID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audit logs",
		})
	}

	return c.JSON(logs)
}

func (h *AuditLogHandler) GetLog(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	id := c.Params("id")

	var log models.AuditLog
	err := h.auditRepo.GetByID(id, shopID, &log)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Audit log not found",
		})
	}

	return c.JSON(log)
}

func (h *AuditLogHandler) GetByUser(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	userID := c.Params("userID")

	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}

	logs, err := h.auditRepo.GetByShopAndUser(shopID, userID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user audit logs",
		})
	}

	return c.JSON(logs)
}

func (h *AuditLogHandler) GetByEntity(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	entityType := c.Params("type")
	entityID := c.Params("id")

	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}

	logs, err := h.auditRepo.GetByEntityAndShop(shopID, entityType, entityID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch entity audit logs",
		})
	}

	return c.JSON(logs)
}

func (h *AuditLogHandler) GetByAction(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	action := c.Params("action")

	limit := c.QueryInt("limit", 50)
	if limit > 100 {
		limit = 100
	}

	logs, err := h.auditRepo.GetByAction(shopID, action, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch action audit logs",
		})
	}

	return c.JSON(logs)
}

func (h *AuditLogHandler) GetStatsSummary(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid start_date format. Use YYYY-MM-DD",
			})
		}
	} else {
		start = time.Now().AddDate(0, 0, -30)
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid end_date format. Use YYYY-MM-DD",
			})
		}
	} else {
		end = time.Now()
	}

	logs, err := h.auditRepo.GetByDateRange(shopID, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audit logs for stats",
		})
	}

	actionCounts := make(map[string]int)
	userCounts := make(map[string]int)
	for _, log := range logs {
		actionCounts[log.Action]++
		userKey := string(log.UserType) + ":" + string(rune(log.UserID))
		userCounts[userKey]++
	}

	return c.JSON(fiber.Map{
		"total_logs": len(logs),
		"by_action":  actionCounts,
		"by_user":    userCounts,
		"start_date": start.Format("2006-01-02"),
		"end_date":   end.Format("2006-01-02"),
	})
}
