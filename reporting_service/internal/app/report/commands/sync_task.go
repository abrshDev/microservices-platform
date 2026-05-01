package commands

import (
	"context"

	"github.com/abrshDev/reporting-service/internal/domain/repositories"
)

type SyncTaskHandler struct {
	repo repositories.SummaryRepo
}

func NewSyncTaskHandler(repo repositories.SummaryRepo) *SyncTaskHandler {
	return &SyncTaskHandler{repo: repo}
}

func (h *SyncTaskHandler) Execute(ctx context.Context, userID string, tenantID uint64, action string) error {
	change := 0
	if action == "TASK_CREATED" {
		change = 1
	} else if action == "TASK_DELETED" {
		change = -1
	}

	if change == 0 {
		return nil
	}

	return h.repo.UpdateWithAudit(userID, tenantID, change, action)
}
