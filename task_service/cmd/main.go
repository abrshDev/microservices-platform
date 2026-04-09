package main

import (
	"log"
	"os"

	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/abrshDev/task-service/internal/app/task/queries"
	"github.com/abrshDev/task-service/internal/delivery/http"
	"github.com/abrshDev/task-service/internal/delivery/http/handlers"
	"github.com/abrshDev/task-service/internal/infrastructure/config"
	"github.com/abrshDev/task-service/internal/infrastructure/database/postgres"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/abrshDev/task-service/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config.LoadEnv()
	appLogger := logger.NewLogger("task-service")

	// 1. Initialize Infrastructure (DB)
	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 2. Service Discovery Addresses
	userSvcAddr := os.Getenv("USER_SERVICE_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051"
	}

	notSvcAddr := os.Getenv("NOTIFICATION_SERVICE_ADDR")
	if notSvcAddr == "" {
		notSvcAddr = "localhost:50052"
	}

	// 3. Initialize gRPC Clients
	// User Client
	userClient, err := grpc.NewUserClient(userSvcAddr)
	if err != nil {
		log.Fatalf("Failed to initialize User gRPC client: %v", err)
	}
	defer userClient.Close()

	// Notification Client
	notifClient, err := grpc.NewNotificationClient(notSvcAddr, appLogger)
	if err != nil {
		log.Fatalf("Failed to initialize Notification gRPC client: %v", err)
	}

	// 4. Dependency Injection
	taskRepo := postgres.NewTaskRepository(db)

	// Injecting both clients into the Command Handler
	createTaskCmd := commands.NewCreateTaskHandler(taskRepo, userClient, notifClient, appLogger)
	getTaskQuery := queries.NewGetTaskHandler(taskRepo, userClient, appLogger)

	taskHandler := handlers.NewTaskHandler(createTaskCmd, getTaskQuery, appLogger)

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
