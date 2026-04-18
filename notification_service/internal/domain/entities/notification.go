package entities

import "fmt"

type Notification struct {
	UserID  string
	Message string
	Status  string
}

// NewTaskNotification prepares the domain entity for the handler
func NewTaskNotification(userID, taskTitle string) *Notification {
	return &Notification{
		UserID:  userID,
		Message: fmt.Sprintf("DUMMY ALERT: New task created -> %s", taskTitle),
		Status:  "PENDING_DISPATCH",
	}
}
