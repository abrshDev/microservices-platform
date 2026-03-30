package commands

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
	domErrors "github.com/abrshDev/user-service/internal/domain/errors"
	"github.com/abrshDev/user-service/internal/domain/repositories"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"` // 72 is Bcrypt's max limit
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	// Check if user exists
	existing, err := h.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return domErrors.ErrEmailAlreadyInUse
	}

	user := &entities.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	return h.repo.Create(ctx, user)
}
