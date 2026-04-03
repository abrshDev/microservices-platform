package commands

import (
	"context"
	"fmt"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/google/uuid"
)

type CreateTaskCommand struct {
	UserID      string
	Title       string
	Description string
}

type CreateTaskHandler struct {
	repo       repositories.TaskRepository
	userClient *grpc.UserClient
}

func NewCreateTaskHandler(repo repositories.TaskRepository, userClient *grpc.UserClient) *CreateTaskHandler {
	return &CreateTaskHandler{
		repo:       repo,
		userClient: userClient,
	}
}

func (h *CreateTaskHandler) Execute(ctx context.Context, cmd CreateTaskCommand) error {
	// 1. Verify user exists via gRPC
	exists, err := h.userClient.CheckUserExists(ctx, cmd.UserID)
	if err != nil {
		// This is a system failure (e.g., User Service is down)
		return fmt.Errorf("internal validation error: %w", err)
	}
	if !exists {
		// This is a validation failure (User provided a bad ID)
		return fmt.Errorf("user %s not found", cmd.UserID)
	}
	parsedUserID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return fmt.Errorf("invalid user uuid format: %w", err)
	}
	// 2. Create the task entity
	task := &entities.Task{
		UserID:      parsedUserID,
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      "PENDING",
	}

	// 3. Save to local Task DB
	return h.repo.Create(ctx, task)
}
