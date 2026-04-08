package commands

import (
	"context"
	"log/slog"
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

	h.logger.Info("Notification dispatched",
		"user_id", cmd.UserID,
		"msg", cmd.Message,
		"type", cmd.Type)
	return nil
}
