package commands

import (
	"context"
	"fmt"
	"log/slog"

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
	repo        repositories.TaskRepository
	userClient  *grpc.UserClient
	notifClient *grpc.NotificationClient // Fixed typo here (Clinet -> Client)
	logger      *slog.Logger
}

func NewCreateTaskHandler(repo repositories.TaskRepository, userClient *grpc.UserClient, notifClient *grpc.NotificationClient, logger *slog.Logger) *CreateTaskHandler {
	return &CreateTaskHandler{
		repo:        repo,
		userClient:  userClient,
		notifClient: notifClient,
		logger:      logger,
	}
}

func (h *CreateTaskHandler) Execute(ctx context.Context, cmd CreateTaskCommand) (*user.UserResponse, error) {
	h.logger.Info("executing create task command",
		slog.String("user_id", cmd.UserID),
		slog.String("title", cmd.Title),
	)

	// 1. Verify user exists and get data via gRPC
	userData, err := h.userClient.GetUser(ctx, cmd.UserID)
	if err != nil {
		h.logger.Error("failed to verify user via gRPC",
			slog.String("user_id", cmd.UserID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("internal validation error: %w", err)
	}

	if userData == nil {
		h.logger.Warn("user not found during task creation",
			slog.String("user_id", cmd.UserID),
		)
		return nil, fmt.Errorf("user %s not found", cmd.UserID)
	}

	// 2. Parse UUID
	parsedUserID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		h.logger.Error("failed to parse user uuid",
			slog.String("user_id", cmd.UserID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("invalid user uuid format: %w", err)
	}

	// 3. Create the task entity
	task := &entities.Task{
		ID:          uuid.New(),
		UserID:      parsedUserID,
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      "PENDING",
	}

	// 4. Save to repository
	if err := h.repo.Create(ctx, task); err != nil {
		h.logger.Error("failed to save task to database",
			slog.String("task_id", task.ID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// 5. Trigger Notification (Added Logic)
	// We run this in a background goroutine so it doesn't block the API response.
	go h.notifClient.SendTaskNotification(context.Background(), cmd.UserID, cmd.Title)

	h.logger.Info("task created successfully",
		slog.String("task_id", task.ID.String()),
		slog.String("assigned_to", userData.Username),
	)

	return userData, nil
}
