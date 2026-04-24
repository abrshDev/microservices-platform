package queries

import (
	"context"

	"github.com/abrshDev/reporting-service/internal/domain/entities"
	"github.com/abrshDev/reporting-service/internal/domain/repositories"
)

type GetSummaryQuery struct {
	repo repositories.SummaryRepo
}

func NewGetSummaryQuery(repo repositories.SummaryRepo) *GetSummaryQuery {
	return &GetSummaryQuery{repo: repo}
}

func (q *GetSummaryQuery) Execute(ctx context.Context, userID string, tenantID uint64) (*entities.UserTaskSummary, error) {
	return q.repo.GetSummary(userID, tenantID)
}
