package commands

import (
	"context"
	"errors"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"github.com/abrshDev/user-service/internal/domain/repositories"
)

// CreateUserRequest defines the data needed to create a user
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
	// 1. Validation (Business Rules)
	if req.Username == "" || req.Email == "" {
		return errors.New("username and email cannot be empty")
	}

	// 2. Convert Request to Domain Entity
	user := &entities.User{
		Username: req.Username,
		Email:    req.Email,
	}

	// 3. Use the Repository to save
	return h.repo.Create(ctx, user)
}
