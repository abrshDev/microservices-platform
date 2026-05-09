package postgres

import (
	"context"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"gorm.io/gorm"
)

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) repositories.OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Create(ctx context.Context, event *entities.OutboxEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *outboxRepository) FetchPendingEvents(ctx context.Context, limit int) ([]entities.OutboxEvent, error) {
	var events []entities.OutboxEvent
	err := r.db.WithContext(ctx).
		Where("status = ?", "PENDING").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *outboxRepository) MarkAsSent(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entities.OutboxEvent{}).
		Where("id = ?", id).
		Update("status", "SENT").Error
}
