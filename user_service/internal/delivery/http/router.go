package http

import (
	"github.com/abrshDev/user-service/internal/delivery/http/handlers"
	"github.com/abrshDev/user-service/internal/delivery/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, handler *handlers.UserHandler) {
	api := app.Group("/api/v1")

	// Public
	api.Post("/login", handler.Login)
	api.Post("/users", handler.CreateUser)

	// Protected
	api.Get("/users/:id", middleware.Protected(), handler.GetUser)
	api.Delete("/users/:id", middleware.Protected(), handler.DeleteUser)
}
