package commands

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"github.com/abrshDev/user-service/internal/domain/repositories"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
}

type CreateUserHandler struct {
	repo repositories.UserRepository
}

func NewCreateUserHandler(repo repositories.UserRepository) *CreateUserHandler {
	return &CreateUserHandler{repo: repo}
}

func (h *CreateUserHandler) Execute(ctx context.Context, req CreateUserRequest) error {
	if err := validate.Struct(req); err != nil {
		return err
	}

	user := &entities.User{
		Username: req.Username,
		Email:    req.Email,
	}

	return h.repo.Create(ctx, user)
}
