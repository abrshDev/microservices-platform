package queries

import (
	"context"
	"errors"
	"os"
	"time"

	domErrors "github.com/abrshDev/user-service/internal/domain/errors"
	"github.com/abrshDev/user-service/internal/domain/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginHandler struct {
	repo      repositories.UserRepository
	jwtSecret []byte
}

func NewLoginHandler(repo repositories.UserRepository) *LoginHandler {
	secret := os.Getenv("JWT_SECRET")
	/* if secret == "" {
		secret = "dev_secret_key_change_me"
	} */
	return &LoginHandler{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

func (h *LoginHandler) Execute(ctx context.Context, req LoginRequest) (string, error) {
	user, err := h.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domErrors.ErrUserNotFound) {
			return "", errors.New("invalid email or password")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}
