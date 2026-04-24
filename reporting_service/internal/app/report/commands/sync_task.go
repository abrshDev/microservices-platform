package commands

import (
	"context"
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/entities"
	"github.com/abrshDev/reporting-service/internal/domain/repositories"
)

type SyncTaskHandler struct {
	repo repositories.SummaryRepo
}

func NewSyncTaskHandler(repo repositories.SummaryRepo) *SyncTaskHandler {
	return &SyncTaskHandler{repo: repo}
}

func (h *SyncTaskHandler) Execute(ctx context.Context, userID string, tenantID uint64) error {
	summary := entities.UserTaskSummary{
		UserID:     userID,
		TenantID:   tenantID,
		TotalTasks: 1, // Set to 1 for the initial atomic increment
		UpdatedAt:  time.Now(),
	}
	return h.repo.UpsertSummary(summary)
}
