package http

import (
	"github.com/abrshDev/reporting-service/internal/delivery/http/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handlers.ReportHandler) {
	api := app.Group("/api/v1")

	api.Get("/reports/:tenantId/:userId", h.GetUserSummary)
}
