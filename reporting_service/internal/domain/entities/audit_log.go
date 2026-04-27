package entities

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        string    `gorm:"column:user_id;index" json:"user_id"`
	TenantID      uint64    `gorm:"column:tenant_id;index" json:"tenant_id"`
	ActionType    string    `gorm:"column:action_type" json:"action_type"`
	PreviousTotal int       `gorm:"column:previous_total" json:"previous_total"`
	NewTotal      int       `gorm:"column:new_total" json:"new_total"`
	CreatedAt     time.Time `json:"created_at"`
}
