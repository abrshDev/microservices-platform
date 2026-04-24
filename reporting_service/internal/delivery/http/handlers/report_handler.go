package handlers

import (
	"strconv"

	"github.com/abrshDev/reporting-service/internal/app/report/queries"
	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	query *queries.GetSummaryQuery
}

func NewReportHandler(query *queries.GetSummaryQuery) *ReportHandler {
	return &ReportHandler{query: query}
}

func (h *ReportHandler) GetUserSummary(c *fiber.Ctx) error {

	tID, _ := strconv.ParseUint(c.Params("tenantId"), 10, 64)

	uID := c.Params("userId")

	summary, err := h.query.Execute(c.Context(), uID, tID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Summary not found"})
	}

	return c.JSON(summary)
}
