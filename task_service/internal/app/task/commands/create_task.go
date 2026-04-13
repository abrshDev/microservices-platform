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
	producer   *kafka.EventProducer // Replaced notifClient
	logger     *slog.Logger
}

func NewCreateTaskHandler(repo repositories.TaskRepository, userClient *grpc.UserClient, producer *kafka.EventProducer, logger *slog.Logger) *CreateTaskHandler {
	return &CreateTaskHandler{
		repo:       repo,
		userClient: userClient,
		producer:   producer, // Replaced notifClient
		logger:     logger,
	}
}

func (h *CreateTaskHandler) Execute(ctx context.Context, cmd CreateTaskCommand) (*user.UserResponse, error) {
	h.logger.Info("executing create task command", slog.String("user_id", cmd.UserID))

	// 1. gRPC: Verify user exists (Keep this synchronous!)
	userData, err := h.userClient.GetUser(ctx, cmd.UserID)
	if err != nil || userData == nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	parsedUserID, _ := uuid.Parse(cmd.UserID)

	// 2. Domain: Create task entity
	task := &entities.Task{
		ID:          uuid.New(),
		UserID:      parsedUserID,
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      "PENDING",
	}

	// 3. Infrastructure: Save to Postgres
	if err := h.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	// 4. Kafka: Publish Event (Replacement for goroutine)
	event := events.TaskCreatedEvent{
		TaskID:      task.ID,
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
	}

	// We publish the event. If Kafka is down, this will return an error,
	// and we can decide whether to fail the whole request or log it.
	if err := h.producer.PublishTaskCreated(ctx, event); err != nil {
		h.logger.Error("failed to publish task created event", slog.String("error", err.Error()))
		// We don't necessarily want to fail the user request if the DB save worked,
		// but we log it heavily for manual fixing.
	}

	return userData, nil
}
