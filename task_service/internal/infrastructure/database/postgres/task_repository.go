package postgres

import (
	"context"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// taskRepository is lowercase (private) to enforce using the Constructor
type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository returns the Interface from the Domain layer
func NewTaskRepository(db *gorm.DB) repositories.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *entities.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Task, error) {
	var task entities.Task
	if err := r.db.WithContext(ctx).First(&task, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.TaskStatus) error {
	return r.db.WithContext(ctx).Model(&entities.Task{}).Where("id = ?", id).Update("status", status).Error
}

func (r *taskRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Task, error) {
	var tasks []entities.Task
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Task{}, "id = ?", id).Error
}
