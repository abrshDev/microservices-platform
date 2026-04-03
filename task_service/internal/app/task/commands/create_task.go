package commands

import (
	"context"
	"fmt"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/abrshDev/task-service/internal/transport/grpc/proto/user"
	"github.com/google/uuid"
)

type CreateTaskCommand struct {
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
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

func (h *CreateTaskHandler) Execute(ctx context.Context, cmd CreateTaskCommand) (*user.UserResponse, error) {
	userData, err := h.userClient.GetUser(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("internal validation error: %w", err)
	}
	if userData == nil {
		return nil, fmt.Errorf("user %s not found", cmd.UserID)
	}

	parsedUserID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid format: %w", err)
	}

	task := &entities.Task{
		ID:          uuid.New(),
		UserID:      parsedUserID,
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      "PENDING",
	}

	if err := h.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return userData, nil
}
