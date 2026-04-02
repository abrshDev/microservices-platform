package handlers

import (
	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	// 1. Define local request structure
	type request struct {
		UserID      string `json:"user_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	// 2. Validate and convert UUID
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	// 3. Map to Command DTO
	commandReq := commands.CreateTaskRequest{
		UserID:      userUUID,
		Title:       req.Title,
		Description: req.Description,
	}

	// 4. Execute Command
	task, err := h.createTaskHandler.Execute(c.Context(), commandReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}
