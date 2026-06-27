package http

import (
	"github.com/abrshDev/user-service/internal/delivery/http/handlers"
	"github.com/abrshDev/user-service/internal/delivery/http/middleware"
	"github.com/gofiber/fiber/v2"
)

// internal/delivery/http/router.go
func SetupRoutes(
	app *fiber.App,
	handler *handlers.UserHandler,
	internalHandler *handlers.InternalUserHandler,
) {
	api := app.Group("/api/v1")

	// Public — for end users
	api.Post("/login", handler.Login)
	api.Post("/users", handler.CreateUser)
	api.Get("/users", middleware.Protected(), handler.ListUser)
	api.Get("/users/:id", middleware.Protected(), handler.GetUser)
	api.Delete("/users/:id", middleware.Protected(), handler.DeleteUser)

	// Internal — for other services only
	internal := api.Group("/internal", middleware.InternalAuth())
	internal.Get("/users/:id", internalHandler.GetUserByID)
	internal.Get("/users/:id/exists", internalHandler.CheckUserExists)
}
