package commands

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/repositories"
)

type DeleteUserHandler struct {
	repo repositories.UserRepository
}

func NewDeleteUserHandler(repo repositories.UserRepository) *DeleteUserHandler {
	return &DeleteUserHandler{repo: repo}
}

func (h *DeleteUserHandler) Execute(ctx context.Context, id string) error {

	return h.repo.Delete(ctx, id)
}
