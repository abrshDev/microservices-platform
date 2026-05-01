package commands

import (
	"context"

	"github.com/abrshDev/reporting-service/internal/domain/repositories"
)

type SyncUserHandler struct {
	repo repositories.SummaryRepo
}

func NewSyncUserHandler(repo repositories.SummaryRepo) *SyncUserHandler {
	return &SyncUserHandler{repo: repo}
}

func (h *SyncUserHandler) Execute(ctx context.Context, userID string, tenantID uint64) error {
	return h.repo.UpdateWithAudit(userID, tenantID, 0, "USER_REGISTERED")
}
