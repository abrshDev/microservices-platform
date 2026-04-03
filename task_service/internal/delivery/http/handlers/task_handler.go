package handlers

import (
	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/gofiber/fiber/v2"
)

type TaskHandler struct {
	createTaskHandler *commands.CreateTaskHandler
}

func NewTaskHandler(createTaskHandler *commands.CreateTaskHandler) *TaskHandler {
	return &TaskHandler{
		createTaskHandler: createTaskHandler,
	}
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {

	var cmd commands.CreateTaskCommand
	if err := c.BodyParser(&cmd); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON request body",
		})
	}

	// 2. Validation (Basic Field Check)
	if cmd.UserID == "" || cmd.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and title are required fields",
		})
	}

	if err := h.createTaskHandler.Execute(c.UserContext(), cmd); err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 4. Return Success
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Task created successfully",
	})
}
