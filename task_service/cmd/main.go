package main

import (
	"log"
	"os"

	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/abrshDev/task-service/internal/delivery/http"
	"github.com/abrshDev/task-service/internal/delivery/http/handlers"
	"github.com/abrshDev/task-service/internal/infrastructure/config"
	"github.com/abrshDev/task-service/internal/infrastructure/database/postgres"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Separate the loading logic
	config.LoadEnv()

	// 2. Initialize Infrastructure (DB)
	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 3. Dependency Injection (The "Chain")
	taskRepo := postgres.NewTaskRepository(db)
	createTaskCmd := commands.NewCreateTaskHandler(taskRepo)
	taskHandler := handlers.NewTaskHandler(createTaskCmd)

	// 4. Setup Server
	app := fiber.New()
	http.SetupRoutes(app, taskHandler)

	// 5. Run
	port := os.Getenv("TASK_APP_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Task Service is starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Fiber failed to start: %v", err)
	}
}
