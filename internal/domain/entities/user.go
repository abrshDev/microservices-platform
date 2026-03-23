package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"primaryKey;type:uuid" json:"id"`
	Username  string         `gorm:"unique;not null" json:"username"`
	Email     string         `gorm:"unique;not null" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Enables soft deletes
}

// BeforeCreate is a GORM hook to auto-generate a UUID
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}
