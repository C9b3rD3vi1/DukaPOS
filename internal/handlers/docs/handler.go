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

func (h *Handler) SwaggerUI(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DukaPOS API Documentation</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.10.5/swagger-ui.min.css">
    <style>
        body { margin: 0; }
        .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.10.5/swagger-ui-bundle.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.10.5/swagger-ui-standalone-preset.min.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/api/docs/json",
                dom_id: "#swagger-ui",
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout",
                docExpansion: "list",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true,
            });
            window.ui = ui;
        };
    </script>
</body>
</html>`)
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	docsGroup := app.Group("/api/docs")
	docsGroup.Get("/json", h.OpenAPIJSON)
	docsGroup.Get("/yaml", h.OpenAPIYAML)
	docsGroup.Get("/markdown", h.Markdown)
	docsGroup.Get("/", h.SwaggerUI)
	docsGroup.Get("/ui", h.SwaggerUI)
}
