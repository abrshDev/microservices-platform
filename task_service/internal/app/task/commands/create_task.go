package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/domain/events"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/abrshDev/task-service/internal/infrastructure/kafka"
	"github.com/abrshDev/task-service/internal/infrastructure/temporal"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type CreateTaskCommand struct {
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateTaskHandler struct {
	repo           repositories.TaskRepository
	userClient     *grpc.UserClient
	producer       *kafka.EventProducer
	logger         *slog.Logger
	temporalConfig temporal.Config
	temporalClient client.Client
}

func NewCreateTaskHandler(repo repositories.TaskRepository, userClient *grpc.UserClient, producer *kafka.EventProducer, logger *slog.Logger, temporalconfig temporal.Config, temporalclient client.Client) *CreateTaskHandler {
	return &CreateTaskHandler{
		repo:           repo,
		userClient:     userClient,
		producer:       producer,
		logger:         logger,
		temporalConfig: temporalconfig,
		temporalClient: temporalclient,
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
	if statusResp.Role != "admin" && statusResp.Role != "user" {
		h.logger.Error("insufficient permissions")
		return nil, fmt.Errorf("insufficient permissions: %w", err)
	}
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

	correlationID, _ := ctx.Value("correlation_id").(string)

	task := &entities.Task{
		ID:          uuid.New(),
		UserID:      parsedUserID,
		TenantID:    uint64(userData.TenantId),
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      "PENDING",
	}

	event := events.TaskCreatedEvent{
		CorrelationID: correlationID,
		TaskID:        task.ID,
		UserID:        task.UserID,
		TenantID:      task.TenantID,
		Title:         task.Title,
		Description:   task.Description,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal outbox payload",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	outboxEvent := &entities.OutboxEvent{
		ID:        uuid.New(),
		EventType: "TASK_CREATED",
		Payload:   string(payload),
		Status:    "PENDING",
	}

	if err := h.repo.CreateTaskWithOutbox(ctx, task, outboxEvent); err != nil {
		h.logger.Error("failed to save task and outbox event",
			slog.String("task_id", task.ID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	h.logger.Info("task created successfully",
		slog.String("task_id", task.ID.String()),
		slog.String("assigned_to", userData.Username),
	)

	return task, nil
}

func (h *CreateTaskHandler) ExecuteWithTemporal(ctx context.Context, cmd CreateTaskCommand) (*entities.Task, error) {
	workflowID := "task-" + uuid.New().String()

	workflowInput := temporal.CreateTaskInput{
		UserID:      cmd.UserID,
		Title:       cmd.Title,
		Description: cmd.Description,
	}
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: h.temporalConfig.TaskQueue,
	}
	workflowRun, err := h.temporalClient.ExecuteWorkflow(ctx, workflowOptions, temporal.CreateTaskWorkflow, workflowInput)
	if err != nil {
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	// Return a task reference immediately - the workflow runs in background
	task := &entities.Task{
		ID:          uuid.MustParse(workflowID),
		UserID:      uuid.MustParse(cmd.UserID),
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      "PROCESSING",
	}

	h.logger.Info("task workflow started",
		"workflow_id", workflowID,
		"run_id", workflowRun.GetRunID(),
	)

	return task, nil

}
