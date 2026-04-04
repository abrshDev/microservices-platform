package main

import (
	"log"
	"os"

	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/abrshDev/task-service/internal/delivery/http"
	"github.com/abrshDev/task-service/internal/delivery/http/handlers"
	"github.com/abrshDev/task-service/internal/infrastructure/config"
	"github.com/abrshDev/task-service/internal/infrastructure/database/postgres"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc" // 1. Import your new gRPC infra
	"github.com/abrshDev/task-service/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config.LoadEnv()

	// 2. Initialize Infrastructure (DB)
	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 3. Initialize gRPC Client (The "Bridge" to User Service)
	// Pull the address from environment (e.g., "user-service:50051")
	userSvcAddr := os.Getenv("USER_SERVICE_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051"
	}

	userClient, err := grpc.NewUserClient(userSvcAddr)
	if err != nil {
		log.Fatalf("Failed to initialize User gRPC client: %v", err)
	}
	defer userClient.Close() // Clean up the connection on shutdown

	// 4. Dependency Injection (The "Chain")
	taskRepo := postgres.NewTaskRepository(db)
	appLogger := logger.NewLogger("task-service")
	createTaskCmd := commands.NewCreateTaskHandler(taskRepo, userClient, appLogger)

	taskHandler := handlers.NewTaskHandler(createTaskCmd)

	// 5. Setup Server
	app := fiber.New()
	http.SetupRoutes(app, taskHandler)

	port := os.Getenv("TASK_APP_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Task Service is starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Fiber failed to start: %v", err)
	}
}
