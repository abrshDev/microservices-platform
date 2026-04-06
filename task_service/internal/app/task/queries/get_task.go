package queries

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/google/uuid"
)

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
	userClient *grpc.UserClient
	logger     *slog.Logger
}

func NewGetTaskHandler(repo repositories.TaskRepository, userclient *grpc.UserClient, logger *slog.Logger) *GetTaskHandler {
	return &GetTaskHandler{
		repo:       repo,
		userClient: userclient,
		logger:     logger,
	}
}
func (h *GetTaskHandler) Execute(ctx context.Context, query GetTaskQuery) (*GetTaskResponse, error) {
	h.logger.Info("fetching task details", slog.String("task_id", query.ID))
	taskUUID, err := uuid.Parse(query.ID)
	if err != nil {
		h.logger.Warn("invalid uuid provided", slog.String("id", query.ID))
		return nil, fmt.Errorf("invalid task uuid format")
	}
	task, err := h.repo.GetByID(ctx, taskUUID)
	if err != nil {
		h.logger.Error("database error fetching task", slog.String("error", err.Error()))
		return nil, err
	}

	if task == nil {
		return nil, nil // Let the handler turn this into a 404
	}
	userDetail := UserDetail{
		ID:   task.UserID.String(),
		Name: "Unknown",
	}

	userData, err := h.userClient.GetUser(ctx, task.UserID.String())
	if err != nil {
		// : Graceful Degradation
		// We log the error but still return the task data.
		// A down User service shouldn't break the Task view.
		h.logger.Warn("could not fetch user metadata via gRPC",
			slog.String("user_id", task.UserID.String()),
			slog.String("error", err.Error()),
		)
	} else if userData != nil {
		userDetail.Name = userData.Username
		userDetail.Email = userData.Email
	}

	// 4. Return merged response
	return &GetTaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		User:        userDetail,
	}, nil
}
