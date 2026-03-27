package http

import (
	"github.com/abrshDev/user-service/internal/delivery/http/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, userHandler *handlers.UserHandler) {
	api := app.Group("/api/v1")

	// User Routes
	api.Get("/users/:id", userHandler.GetUser)
	api.Post("/users", userHandler.CreateUser)

}
