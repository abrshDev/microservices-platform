package postgres

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"github.com/abrshDev/user-service/internal/domain/repositories" // Import the interface
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// Note: This returns the INTERFACE from the domain
func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
