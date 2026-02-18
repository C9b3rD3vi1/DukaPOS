package docshandler

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/docs"
	"github.com/gofiber/fiber/v2"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) OpenAPIJSON(c *fiber.Ctx) error {
	json, err := docs.GenerateDocsJSON()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	c.Set("Content-Type", "application/json")
	return c.SendString(json)
}

func (h *Handler) OpenAPIYAML(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "YAML format not yet implemented",
		"hint":  "Use /api/docs/json for JSON format",
	})
}

func (h *Handler) Markdown(c *fiber.Ctx) error {
	md := docs.Markdown()
	c.Set("Content-Type", "text/markdown")
	return c.SendString(md)
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	docsGroup := app.Group("/api/docs")
	docsGroup.Get("/json", h.OpenAPIJSON)
	docsGroup.Get("/yaml", h.OpenAPIYAML)
	docsGroup.Get("/markdown", h.Markdown)
	docsGroup.Get("/", h.Markdown)
}
