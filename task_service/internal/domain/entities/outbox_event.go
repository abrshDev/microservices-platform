package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxEvent struct {
	ID        uuid.UUID `json:"id"`
	EventType string    `json:"event_type"`
	Payload   string    `json:"payload"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (e *OutboxEvent) BeforeCreate(tx *gorm.DB) error {
	e.ID = uuid.New()
	return nil
}
