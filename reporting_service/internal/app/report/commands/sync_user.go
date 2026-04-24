package commands

import (
	"context"
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/entities"
	"github.com/abrshDev/reporting-service/internal/domain/repositories"
)

type SyncUserHandler struct {
	repo repositories.SummaryRepo
}

func NewSyncUserHandler(repo repositories.SummaryRepo) *SyncUserHandler {
	return &SyncUserHandler{repo: repo}
}

func (h *SyncUserHandler) Execute(ctx context.Context, userID string, tenantID uint64) error {
	summary := entities.UserTaskSummary{
		UserID:    userID,
		TenantID:  tenantID,
		UpdatedAt: time.Now(),
	}
	return h.repo.UpsertSummary(summary)
}
