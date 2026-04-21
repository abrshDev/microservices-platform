package postgres

import (
	"github.com/abrshDev/reporting-service/internal/domain/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SummaryRepository struct {
	db *gorm.DB
}

func NewSummaryRepository(db *gorm.DB) *SummaryRepository {
	return &SummaryRepository{db: db}
}

func (r *SummaryRepository) UpsertSummary(summary entities.UserTaskSummary) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "tenant_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_tasks": gorm.Expr("total_tasks + 1"),
			"updated_at":  summary.UpdatedAt,
		}),
	}).Create(&summary).Error
}
