package repositories

import (
	"context"

	"github.com/abrshDev/reporting-service/internal/domain/entities"
)

type SummaryRepo interface {
	UpdateWithAudit(userID string, tenantID uint64, change int, actionType string) error
	InsertIfNotExist(ctx context.Context, eventID string) (bool, error)
	UpdateStatus(ctx context.Context, eventID string, status string) error
	GetSummary(userID string, tenantID uint64) (*entities.UserTaskSummary, error)
}
