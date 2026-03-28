package handlers

import (
	"github.com/abrshDev/user-service/internal/app/user/commands"
	"github.com/abrshDev/user-service/internal/app/user/queries"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	createHandler *commands.CreateUserHandler
	getHandler    *queries.GetUserHandler
	DeleteHandler *commands.DeleteUserHandler
}

func NewUserHandler(c *commands.CreateUserHandler, q *queries.GetUserHandler, d *commands.DeleteUserHandler) *UserHandler {
	return &UserHandler{
		createHandler: c,
		getHandler:    q,
		DeleteHandler: d,
	}
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

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id") // Keep it as a string

	query := queries.GetUserQuery{ID: id}
	user, err := h.getHandler.Execute(c.Context(), query)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.DeleteHandler.Execute(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	return c.Status(200).JSON(fiber.Map{"message": "soft deleted successfully"})
}
