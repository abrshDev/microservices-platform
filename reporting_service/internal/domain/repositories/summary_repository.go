package repositories

import "github.com/abrshDev/reporting-service/internal/domain/entities"

type SummaryRepo interface {
	UpsertSummary(summary entities.UserTaskSummary) error
	GetSummary(userID uint, tenantID uint) (*entities.UserTaskSummary, error)
}
