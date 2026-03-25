package commands

import (
	"context"
	"errors"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"github.com/abrshDev/user-service/internal/domain/repositories"
)

type CreateUserRequest struct {
	Username string
	Email    string
}

type CreateUserHandler struct {
	repo repositories.UserRepository
}

func NewCreateUserHandler(repo repositories.UserRepository) *CreateUserHandler {
	return &CreateUserHandler{repo: repo}
}

func (h *CreateUserHandler) Execute(ctx context.Context, req CreateUserRequest) error {
	if req.Username == "" || req.Email == "" {
		return errors.New("username and email cannot be empty")
	}

	user := &entities.User{
		Username: req.Username,
		Email:    req.Email,
	}

	return h.repo.Create(ctx, user)
}
