package queries

import (
	"context"
	"errors"

	domErrors "github.com/abrshDev/user-service/internal/domain/errors"
	"github.com/abrshDev/user-service/internal/domain/repositories"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginHandler struct {
	repo repositories.UserRepository
}

func NewLoginHandler(repo repositories.UserRepository) *LoginHandler {
	return &LoginHandler{repo: repo}
}
func (h *LoginHandler) Execute(ctx context.Context, req LoginRequest) (string, error) {
	user, err := h.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domErrors.ErrUserNotFound) {
			return "", errors.New("invalid email or password")
		}
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {

		return "", errors.New("invalid email or password")
	}

	return user.ID, nil
}
