package queries

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"github.com/abrshDev/user-service/internal/domain/repositories"
)

type GetUserQuery struct {
	ID string
}

type GetUserHandler struct {
	repo repositories.UserRepository
}

func NewGetUserHandler(repo repositories.UserRepository) *GetUserHandler {
	return &GetUserHandler{repo: repo}
}

func (h *GetUserHandler) Execute(ctx context.Context, query GetUserQuery) (*entities.User, error) {
	return h.repo.GetByID(ctx, query.ID)
}
