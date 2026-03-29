package postgres

import (
	"context"
	"errors"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"github.com/abrshDev/user-service/internal/domain/repositories"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {

	return r.db.WithContext(ctx).Delete(&entities.User{}, "id = ?", id).Error
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
