package handlers

import (
	"github.com/abrshDev/user-service/internal/app/user/commands"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	createHandler *commands.CreateUserHandler
}

func NewUserHandler(c *commands.CreateUserHandler) *UserHandler {
	return &UserHandler{createHandler: c}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req commands.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	err := h.createHandler.Execute(c.Context(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User created"})
}
