package http

import (
	"time"

	"github.com/abrshDev/task-service/internal/delivery/http/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func SetupRoutes(app *fiber.App, taskHandler *handlers.TaskHandler) {
	api := app.Group("/api/v1")
	tasks := api.Group("/tasks")
	tasks.Use(limiter.New(limiter.Config{
		Max:        5,
		Expiration: 60 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Too many requests. Try again later.",
			})
		},
	}))
	tasks.Post("/create", taskHandler.CreateTask)
	tasks.Get("/get/:id", taskHandler.GetTask)
}
