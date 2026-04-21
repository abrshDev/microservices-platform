package kafka

import (
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/entities"
)

type TaskEvent struct {
	UserID    uint      `json:"user_id"`
	TenantID  uint      `json:"tenant_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *TaskEvent) MapToEntity() entities.UserTaskSummary {
	return entities.UserTaskSummary{
		UserID:    e.UserID,
		TenantID:  e.TenantID,
		UpdatedAt: e.Timestamp,
	}
}
