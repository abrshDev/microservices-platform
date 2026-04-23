package entities

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "PENDING"
	StatusRunning   TaskStatus = "RUNNING"
	StatusCompleted TaskStatus = "COMPLETED"
	StatusFailed    TaskStatus = "FAILED"
)

type Task struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Title       string    `gorm:"not null"`
	Description string
	Status      TaskStatus `gorm:"type:varchar(20);default:'PENDING'"`
	TenantID    uint       `json:"tenant_id" gorm:"index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
