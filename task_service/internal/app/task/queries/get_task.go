package queries

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/transport/grpc/proto/user"
	"github.com/google/uuid"
)

type UserClientInterface interface {
	GetUser(ctx context.Context, userID string) (*user.UserResponse, error)
}

type GetTaskQuery struct {
	ID string
}

type GetTaskResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	User        UserDetail `json:"user"`
}

type UserDetail struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GetTaskHandler struct {
	repo       repositories.TaskRepository
	userClient UserClientInterface // Changed from *grpc.UserClient to Interface
	logger     *slog.Logger
}

func NewGetTaskHandler(repo repositories.TaskRepository, userClient UserClientInterface, logger *slog.Logger) *GetTaskHandler {
	return &GetTaskHandler{
		repo:       repo,
		userClient: userClient,
		logger:     logger,
	}
}

func (h *GetTaskHandler) Execute(ctx context.Context, query GetTaskQuery) (*GetTaskResponse, error) {
	taskUUID, err := uuid.Parse(query.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid task uuid format")
	}

	task, err := h.repo.GetByID(ctx, taskUUID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}

	// Default "Unknown" user in case gRPC fails (Graceful Degradation)
	userDetail := UserDetail{
		ID:   task.UserID.String(),
		Name: "Unknown",
	}

	// Call gRPC via the interface
	userData, err := h.userClient.GetUser(ctx, task.UserID.String())
	if err != nil {
		h.logger.Warn("gRPC user fetch failed, using defaults", slog.String("error", err.Error()))
	} else if userData != nil {
		userDetail.Name = userData.Username
		userDetail.Email = userData.Email
	}

	return &GetTaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		User:        userDetail,
	}, nil
}
