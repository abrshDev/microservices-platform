package events

import "github.com/google/uuid"

type TaskCreatedEvent struct {
	CorrelationID string    `json:"correlation_id"`
	TaskID        uuid.UUID `json:"task_id"`
	UserID        uuid.UUID `json:"user_id"`
	TenantID      uint64    `json:"tenant_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Action        string    `json:"action"`
}
