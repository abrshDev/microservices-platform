package repositories

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"gorm.io/gorm"
)

// 1. The Interface (The Rules)
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uint) (*entities.User, error)
}

// 2. The Struct (The Worker)
type userRepository struct {
	db *gorm.DB
}

// 3. The Constructor (The Boss who hires the worker)
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// 4. The Implementation (The actual logic)
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*entities.User, error) {
	var user entities.User
	// Using WithContext(ctx) is best practice for timeouts and tracing
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
