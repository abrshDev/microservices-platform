package repositories

import (
	"context"

	"github.com/abrshDev/task-service/internal/domain/entities"
)

type OutboxRepository interface {
	Create(ctx context.Context, event *entities.OutboxEvent) error
	FetchPendingEvents(ctx context.Context, limit int) ([]entities.OutboxEvent, error)
	MarkAsSent(ctx context.Context, id string) error
}
