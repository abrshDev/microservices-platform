package handlers

import (
	"errors"

	"github.com/abrshDev/user-service/internal/app/user/commands"
	"github.com/abrshDev/user-service/internal/app/user/queries"
	domErrors "github.com/abrshDev/user-service/internal/domain/errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	createHandler *commands.CreateUserHandler
	getHandler    *queries.GetUserHandler
	DeleteHandler *commands.DeleteUserHandler
	LoginHandler  *queries.LoginHandler
}
type InternalUserHandler struct {
}

func (h *InternalUserHandler) GetUserByID(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

func (h *InternalUserHandler) CheckUserExists(c *fiber.Ctx) error {
	return c.SendStatus(200)
}
func NewInternalUserHandler() *InternalUserHandler {
	return &InternalUserHandler{}
}
func NewUserHandler(c *commands.CreateUserHandler, q *queries.GetUserHandler, d *commands.DeleteUserHandler, l *queries.LoginHandler) *UserHandler {
	return &UserHandler{
		createHandler: c,
		getHandler:    q,
		DeleteHandler: d,
		LoginHandler:  l,
	}
}
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req commands.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	err := h.createHandler.Execute(c.Context(), req)
	if err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {

			return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
		}

		if errors.Is(err, domErrors.ErrEmailAlreadyInUse) {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User created successfully"})
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id is required"})
	}
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
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req queries.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	accessToken, err := h.LoginHandler.Execute(c.Context(), req)
	if err != nil {

		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"access_token": accessToken,
		"message":      "Login successful",
	})
}

func (h *UserHandler) ListUser(c *fiber.Ctx) error {
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100
	}

	return c.Status(200).JSON(fiber.Map{
		"status": status,
		"page":   page,
		"limit":  limit,
		"data":   []string{}, // hardcoded placeholder for now
	})
}
