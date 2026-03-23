package postgres

import (
	"context"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"gorm.io/gorm"
)

// We use an unexported struct 'userRepository' (lowercase)
// to force people to use the NewUserRepository constructor.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository is clean.
// It returns the interface defined in the domain layer.
func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var user entities.User
	// Using First returns an error if the record is not found
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
