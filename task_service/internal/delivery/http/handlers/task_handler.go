package handlers

import (
	"context"
	"log/slog"

	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/abrshDev/task-service/internal/app/task/queries"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TaskHandler struct {
	createTaskHandler *commands.CreateTaskHandler
	GetTaskHandler    *queries.GetTaskHandler
	logger            *slog.Logger
}

func NewTaskHandler(createTaskHandler *commands.CreateTaskHandler, getTaskHandler *queries.GetTaskHandler, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		createTaskHandler: createTaskHandler,
		GetTaskHandler:    getTaskHandler,
		logger:            logger,
	}
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	// Extract correlation ID from request header, or generate a new one
	correlationID := c.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	// Send correlation ID back to client so they can report it with bugs
	c.Set("X-Correlation-ID", correlationID)

	// Attach correlation ID to context so downstream services can use it
	ctx := context.WithValue(c.Context(), "correlation_id", correlationID)

	// Create a logger that includes the correlation ID in every log line
	logger := h.logger.With(slog.String("correlation_id", correlationID))

	var cmd commands.CreateTaskCommand
	if err := c.BodyParser(&cmd); err != nil {
		logger.Warn("failed to parse request body", slog.String("error", err.Error()))
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	userData, err := h.createTaskHandler.Execute(ctx, cmd)
	if err != nil {
		logger.Error("task creation failed",
			slog.String("user_id", cmd.UserID),
			slog.String("error", err.Error()),
		)
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Task created successfully",
		"data": fiber.Map{
			"user": userData,
		},
	})
}

func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	// Extract or generate correlation ID
	correlationID := c.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	c.Set("X-Correlation-ID", correlationID)
	ctx := context.WithValue(c.Context(), "correlation_id", correlationID)
	logger := h.logger.With(slog.String("correlation_id", correlationID))

	id := c.Params("id")

	task, err := h.GetTaskHandler.Execute(ctx, queries.GetTaskQuery{
		ID: id,
	})

	if err != nil {
		logger.Error("failed to execute get task query",
			slog.String("task_id", id),
			slog.String("error", err.Error()),
		)

		if err.Error() == "invalid task uuid format" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid ID format",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	if task == nil {
		logger.Warn("task lookup returned no results", slog.String("task_id", id))
		return c.Status(404).JSON(fiber.Map{
			"error": "Task not found",
		})
	}

	return c.Status(200).JSON(task)
}
