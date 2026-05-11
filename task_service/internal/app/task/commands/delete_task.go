package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/google/uuid"
)

type DeleteTaskCommand struct {
	TaskID string
	UserID string
}

type DeleteTaskHandler struct {
	repo       repositories.TaskRepository
	userClient *grpc.UserClient
	logger     *slog.Logger
}

func NewDeleteTaskHandler(repo repositories.TaskRepository, userClient *grpc.UserClient, logger *slog.Logger) *DeleteTaskHandler {
	return &DeleteTaskHandler{
		repo:       repo,
		userClient: userClient,
		logger:     logger,
	}
}

func (h *DeleteTaskHandler) Execute(ctx context.Context, cmd DeleteTaskCommand) error {
	taskUUID, err := uuid.Parse(cmd.TaskID)
	if err != nil {
		return fmt.Errorf("invalid task id format")
	}

	statusResp, err := h.userClient.CheckUserStatus(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}

	if !statusResp.IsActive {
		return fmt.Errorf("user is inactive")
	}

	task, err := h.repo.GetByID(ctx, taskUUID)
	if err != nil {
		return fmt.Errorf("task not found")
	}

	if statusResp.Role == "admin" {
		return h.repo.Delete(ctx, taskUUID)
	}

	if statusResp.Role == "user" && task.UserID.String() == cmd.UserID {
		return h.repo.Delete(ctx, taskUUID)
	}

	return fmt.Errorf("forbidden: you do not have permission to delete this task")
}
