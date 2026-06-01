package temporal

import (
	"context"
	"fmt"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/domain/events"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/abrshDev/task-service/internal/infrastructure/kafka"
	"github.com/google/uuid"
)

// Activities holds dependencies needed by workflow activities
type Activities struct {
	TaskRepo   repositories.TaskRepository
	UserClient *grpc.UserClient
	Producer   *kafka.EventProducer
}

// ValidateUserActivity checks if a user exists and is active via gRPC
func (a *Activities) ValidateUserActivity(ctx context.Context, input ValidateUserInput) (*ValidateUserResult, error) {
	userData, err := a.UserClient.GetUser(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	if userData == nil {
		return nil, fmt.Errorf("user %s not found", input.UserID)
	}

	return &ValidateUserResult{
		TenantID: uint64(userData.TenantId),
	}, nil
}

// SaveTaskActivity persists the task to the database
func (a *Activities) SaveTaskActivity(ctx context.Context, input SaveTaskInput) (*SaveTaskResult, error) {
	task := &entities.Task{
		ID:          uuid.New(),
		UserID:      uuid.MustParse(input.UserID),
		TenantID:    input.TenantID,
		Title:       input.Title,
		Description: input.Description,
		Status:      "PENDING",
	}

	if err := a.TaskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return &SaveTaskResult{
		TaskID: task.ID.String(),
	}, nil
}

// PublishEventActivity publishes a task created event to Kafka
func (a *Activities) PublishEventActivity(ctx context.Context, input PublishEventInput) (*PublishEventResult, error) {
	event := events.TaskCreatedEvent{
		TaskID:      uuid.MustParse(input.TaskID),
		UserID:      uuid.MustParse(input.UserID),
		TenantID:    input.TenantID,
		Title:       input.Title,
		Description: input.Description,
	}

	if err := a.Producer.PublishTaskCreated(ctx, input.UserID, event); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return &PublishEventResult{Success: true}, nil
}
