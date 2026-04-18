package postgres

import (
	"context"

	"github.com/abrshDev/notification_service/internal/domain/repositories"
	"gorm.io/gorm"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) repositories.EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) InsertIfNotExist(ctx context.Context, eventID string) (bool, error) {

	result := r.db.WithContext(ctx).Exec(
		"INSERT INTO processed_events (event_id, status) VALUES (?, ?) ON CONFLICT (event_id) DO NOTHING",
		eventID, "PENDING",
	)

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

func (r *eventRepository) UpdateStatus(ctx context.Context, eventID string, status string) error {

	return r.db.WithContext(ctx).Exec(
		"UPDATE processed_events SET status = ? WHERE event_id = ?",
		status, eventID,
	).Error
}
