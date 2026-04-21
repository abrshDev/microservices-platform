package entities

import "time"

type UserTaskSummary struct {
	UserID     uint `gorm:"primaryKey;autoIncrement:false"`
	TenantID   uint `gorm:"primaryKey;autoIncrement:false"`
	TotalTasks int  `gorm:"default:0"`
	UpdatedAt  time.Time
}
