package postgres

import (
	"github.com/abrshDev/reporting-service/internal/domain/entities"
	"github.com/abrshDev/reporting-service/internal/domain/repositories"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SummaryRepository struct {
	db *gorm.DB
}

func NewSummaryRepository(db *gorm.DB) repositories.SummaryRepo {
	return &SummaryRepository{db: db}
}

func (r *SummaryRepository) UpsertSummary(summary entities.UserTaskSummary) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "tenant_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{

			"total_tasks": gorm.Expr("user_task_summaries.total_tasks + 1"),
			"updated_at":  summary.UpdatedAt,
		}),
	}).Create(&summary).Error
}

func (r *SummaryRepository) GetSummary(userID string, tenantID uint64) (*entities.UserTaskSummary, error) {
	var summary entities.UserTaskSummary

	err := r.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}
