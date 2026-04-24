package entities

import "time"

type UserTaskSummary struct {
	UserID     string `gorm:"primaryKey;autoIncrement:false"`
	TenantID   uint64 `gorm:"primaryKey;autoIncrement:false"`
	TotalTasks int    `gorm:"default:0"`
	UpdatedAt  time.Time
}
