package http

import (
	"github.com/abrshDev/reporting-service/internal/delivery/http/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handlers.ReportHandler) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Reporting Service is Online")
	})

	api := app.Group("/api/v1")

	api.Get("/reports/:tenantId/:userId", h.GetUserSummary)
}
