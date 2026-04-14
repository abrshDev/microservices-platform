package commands

import (
	"context"
	"log/slog"

	"github.com/abrshDev/notification_service/internal/enitities"
)

type SendNotificationCommand struct {
	UserID  string
	Message string
	Type    string
}

type SendNotificationHandler struct {
	logger *slog.Logger
}

func NewSendNotificationHandler(logger *slog.Logger) *SendNotificationHandler {
	return &SendNotificationHandler{logger: logger}
}

func (h *SendNotificationHandler) Handle(ctx context.Context, cmd SendNotificationCommand) error {
	n := enitities.NewTaskNotification(cmd.UserID, cmd.Message)

	h.logger.Info("Executing notification delivery",
		"user_id", n.UserID,
		"content", n.Message,
		"status", "SUCCESS",
	)
	return nil
}
