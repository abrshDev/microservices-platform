package commands

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
	domErrors "github.com/abrshDev/user-service/internal/domain/errors"
	"github.com/abrshDev/user-service/internal/domain/repositories"
	"github.com/abrshDev/user-service/internal/infrastructure/kafka"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	// Added TenantID with json tag matching your CURL request ("tenant_id")
	TenantID uint `json:"tenant_id" validate:"required"`
}

type CreateUserHandler struct {
	repo     repositories.UserRepository
	producer *kafka.UserProducer
}

func NewCreateUserHandler(repo repositories.UserRepository, producer *kafka.UserProducer) *CreateUserHandler {
	return &CreateUserHandler{repo: repo, producer: producer}
}

func (h *CreateUserHandler) Execute(ctx context.Context, req CreateUserRequest) error {
	// 1. Validate input
	// This now ensures TenantID is present in the request
	if err := validate.Struct(req); err != nil {
		return err
	}

	// 2. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 3. Check for existing user
	existing, err := h.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return domErrors.ErrEmailAlreadyInUse
	}

	// Map the request data to the User entity
	user := &entities.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),

		TenantID: req.TenantID,
	}

	// 4. Save to User Database
	if err := h.repo.Create(ctx, user); err != nil {
		return err
	}

	// 5. Tell the world (Kafka) a new user was created
	err = h.producer.PublishUserCreated(ctx, user.ID, user.Email, user.TenantID)
	if err != nil {
		println("ERROR: Kafka failed to publish, but user was created:", err.Error())
		return nil
	}
	return nil

}
