package repositories

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id string) (*entities.User, error)
	Delete(ctx context.Context, id string) error
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
}
