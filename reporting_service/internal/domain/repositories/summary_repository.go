package repositories

import "github.com/abrshDev/reporting-service/internal/domain/entities"

type SummaryRepo interface {
	UpdateWithAudit(userID string, tenantID uint64, change int, actionType string) error

	GetSummary(userID string, tenantID uint64) (*entities.UserTaskSummary, error)
}
