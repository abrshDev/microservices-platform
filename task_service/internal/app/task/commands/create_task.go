package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/domain/events"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/abrshDev/task-service/internal/infrastructure/kafka"
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
	producer   *kafka.EventProducer
	logger     *slog.Logger
}

func NewCreateTaskHandler(repo repositories.TaskRepository, userClient *grpc.UserClient, producer *kafka.EventProducer, logger *slog.Logger) *CreateTaskHandler {
	return &CreateTaskHandler{
		repo:       repo,
		userClient: userClient,
		producer:   producer,
		logger:     logger,
	}
}

func (h *CreateTaskHandler) Execute(ctx context.Context, cmd CreateTaskCommand) (*entities.Task, error) {
	parsedUserID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		h.logger.Error("failed to parse user uuid",
			slog.String("user_id", cmd.UserID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("invalid user uuid format: %w", err)
	}

	h.logger.Info("executing create task command",
		slog.String("user_id", cmd.UserID),
		slog.String("title", cmd.Title),
	)

	statusResp, err := h.userClient.CheckUserStatus(ctx, cmd.UserID)
	if err != nil {
		h.logger.Error("gRPC CheckUserStatus failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if !statusResp.IsActive {
		h.logger.Warn("user is inactive", slog.String("user_id", cmd.UserID))
		return nil, fmt.Errorf("cannot create task: user %s is inactive", cmd.UserID)
	}
	h.logger.Info("user is active")

	userData, err := h.userClient.GetUser(ctx, cmd.UserID)
	if err != nil {
		h.logger.Error("failed to verify user via gRPC",
			slog.String("user_id", cmd.UserID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("internal validation error: %w", err)
	}

	if userData == nil {
		h.logger.Warn("user not found during task creation", slog.String("user_id", cmd.UserID))
		return nil, fmt.Errorf("user %s not found", cmd.UserID)
	}

	task := &entities.Task{
		ID:          uuid.New(),
		UserID:      parsedUserID,
		TenantID:    uint64(userData.TenantId),
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      "PENDING",
	}

	if err := h.repo.Create(ctx, task); err != nil {
		h.logger.Error("failed to save task to database",
			slog.String("task_id", task.ID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	event := events.TaskCreatedEvent{
		TaskID:      task.ID,
		UserID:      task.UserID,
		TenantID:    task.TenantID,
		Title:       task.Title,
		Description: task.Description,
	}

	if err := h.producer.PublishTaskCreated(ctx, parsedUserID.String(), event); err != nil {
		h.logger.Error("failed to publish task created event to kafka",
			slog.String("task_id", task.ID.String()),
			slog.String("error", err.Error()),
		)
		return task, fmt.Errorf("task created but failed to publish event: %w", err)
	}

	h.logger.Info("task created successfully and event published",
		slog.String("task_id", task.ID.String()),
		slog.String("assigned_to", userData.Username),
	)

	return task, nil
}
