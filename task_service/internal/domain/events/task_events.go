package events

import "github.com/google/uuid"

type TaskEvent struct {
	TaskId      uuid.UUID `json:"task_id"`
	UserId      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}
