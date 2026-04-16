package queries

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/repositories"
)

type GetUserStatusQuery struct {
	ID string
}
type CheckUserStatusResult struct {
	IsActive bool
	Role     string
}
type CheckUserStatusHandler struct {
	repo repositories.UserRepository
}

func NewCheckUserStatusHandler(repo repositories.UserRepository) *CheckUserStatusHandler {
	return &CheckUserStatusHandler{
		repo: repo,
	}
}

func (h *CheckUserStatusHandler) Execute(ctx context.Context, query GetUserStatusQuery) (*CheckUserStatusResult, error) {
	user, err := h.repo.FindByID(ctx, query.ID)
	if err != nil {
		return nil, err
	}

	return &CheckUserStatusResult{IsActive: user.IsActive,
		Role: user.Role,
	}, nil
}
