package repositories

import "github.com/abrshDev/reporting-service/internal/domain/entities"

type SummaryRepo interface {
	UpsertSummary(summary entities.UserTaskSummary) error
	GetSummary(userID string, tenantID uint64) (*entities.UserTaskSummary, error)
}
