package handlers

import (
	"log/slog"

	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/abrshDev/task-service/internal/app/task/queries"
	"github.com/gofiber/fiber/v2"
)

type TaskHandler struct {
	createTaskHandler *commands.CreateTaskHandler
	GetTaskHandler    *queries.GetTaskHandler
	logger            *slog.Logger
}

// Update the constructor to accept the logger
func NewTaskHandler(createTaskHandler *commands.CreateTaskHandler, getTaskHandler *queries.GetTaskHandler, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		createTaskHandler: createTaskHandler,
		GetTaskHandler:    getTaskHandler,
		logger:            logger,
	}
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	var cmd commands.CreateTaskCommand
	if err := c.BodyParser(&cmd); err != nil {
		h.logger.Warn("failed to parse request body", slog.String("error", err.Error()))
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	userData, err := h.createTaskHandler.Execute(c.Context(), cmd)
	if err != nil {

		h.logger.Error("task creation failed",
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

	id := c.Params("id")

	task, err := h.GetTaskHandler.Execute(c.Context(), queries.GetTaskQuery{
		ID: id,
	})

	if err != nil {
		h.logger.Error("failed to execute get task query",
			slog.String("task_id", id),
			slog.String("error", err.Error()),
		)

		// If the error is specifically about UUID format, return 400
		if err.Error() == "invalid task uuid format" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid ID format",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// . Handle the "Not Found" case
	if task == nil {
		h.logger.Warn("task lookup returned no results", slog.String("task_id", id))
		return c.Status(404).JSON(fiber.Map{
			"error": "Task not found",
		})
	}

	// Return the merged Task + User data
	return c.Status(200).JSON(task)
}
