package postgres

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"

	"gorm.io/gorm"
)

type gormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *gormUserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return &user, err
}
