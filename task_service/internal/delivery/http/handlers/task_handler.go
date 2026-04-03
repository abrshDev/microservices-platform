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
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	userData, err := h.createTaskHandler.Execute(c.Context(), cmd)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Task created successfully",
		"data": fiber.Map{
			"task_title": cmd.Title,
			"user": fiber.Map{
				"name":  userData.Username,
				"email": userData.Email,
			},
		},
	})
}
