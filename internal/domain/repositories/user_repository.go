package repositories

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uint) (*entities.User, error)
}
