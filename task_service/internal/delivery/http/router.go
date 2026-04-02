package http

import (
	"github.com/abrshDev/task-service/internal/delivery/http/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, taskHandler *handlers.TaskHandler) {
	api := app.Group("/api/v1")

	tasks := api.Group("/tasks")
	tasks.Post("/create", taskHandler.CreateTask)

}
