package commands

import (
	"context"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	UserID      uuid.UUID
	Title       string
	Description string
}

type CreateTaskHandler struct {
	repo repositories.TaskRepository
}

func NewCreateTaskHandler(repo repositories.TaskRepository) *CreateTaskHandler {
	return &CreateTaskHandler{repo: repo}
}

func (h *CreateTaskHandler) Execute(ctx context.Context, req CreateTaskRequest) (*entities.Task, error) {
	task := &entities.Task{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Status:      entities.StatusPending,
	}

	if err := h.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}
