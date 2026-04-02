package repositories

import (
	"context"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/google/uuid"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entities.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Task, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.TaskStatus) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
