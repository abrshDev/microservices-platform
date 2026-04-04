package handlers

import (
	"log/slog" // Import slog

	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/gofiber/fiber/v2"
)

type TaskHandler struct {
	createTaskHandler *commands.CreateTaskHandler
	logger            *slog.Logger // Add this line
}

// Update the constructor to accept the logger
func NewTaskHandler(createTaskHandler *commands.CreateTaskHandler, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		createTaskHandler: createTaskHandler,
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
		// Now h.logger will work here!
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
