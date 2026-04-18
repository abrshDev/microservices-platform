package repositories

import "context"

type EventRepository interface {
	InsertIfNotExist(ctx context.Context, eventID string) (bool, error)

	UpdateStatus(ctx context.Context, eventID string, status string) error
}
