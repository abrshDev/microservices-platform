package commands

import (
	"context"
	"log/slog"

	"github.com/abrshDev/notification_service/internal/domain/entities"
	"github.com/abrshDev/notification_service/internal/domain/repositories"
)

type SendNotificationCommand struct {
	TaskID  string
	UserID  string
	Message string
	Type    string
}

type SendNotificationHandler struct {
	logger *slog.Logger
	repo   repositories.EventRepository
}

func NewSendNotificationHandler(logger *slog.Logger, repo repositories.EventRepository) *SendNotificationHandler {
	return &SendNotificationHandler{logger: logger, repo: repo}
}

func (h *SendNotificationHandler) Execute(ctx context.Context, cmd SendNotificationCommand) error {
	// Check if we already processed this event
	isNew, err := h.repo.InsertIfNotExist(ctx, cmd.TaskID)
	if err != nil {
		h.logger.Error("failed to check idempotency", "task_id", cmd.TaskID, "error", err)
		return err
	}

	if !isNew {
		h.logger.Info("duplicate event detected, skipping", "task_id", cmd.TaskID)
		return nil
	}

	// Create notification entity
	n := entities.NewTaskNotification(cmd.UserID, cmd.Message)

	// Dispatch notification
	h.logger.Info("Executing notification delivery",
		"user_id", n.UserID,
		"content", n.Message,
		"status", "SUCCESS",
	)

	// Mark event as completed in database
	return h.repo.UpdateStatus(ctx, cmd.TaskID, "COMPLETED")
}
