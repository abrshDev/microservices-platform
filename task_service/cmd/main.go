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
	"github.com/abrshDev/task-service/internal/infrastructure/kafka" // Added
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

	// 2. Initialize Infrastructure (Kafka)
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:29092" // Use internal docker network address
	}
	taskProducer := kafka.NewEventProducer([]string{kafkaBrokers}, "task-events")

	// 3. Service Discovery Addresses
	userSvcAddr := os.Getenv("USER_SERVICE_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "user-service:50051"
	}

	// 4. Initialize gRPC Clients
	// User Client (Kept for synchronous validation)
	userClient, err := grpc.NewUserClient(userSvcAddr)
	if err != nil {
		log.Fatalf("Failed to initialize User gRPC client: %v", err)
	}
	defer userClient.Close()

	// 5. Dependency Injection
	taskRepo := postgres.NewTaskRepository(db)

	// Injecting UserClient (Sync) and taskProducer (Async)
	createTaskCmd := commands.NewCreateTaskHandler(taskRepo, userClient, taskProducer, appLogger)
	getTaskQuery := queries.NewGetTaskHandler(taskRepo, userClient, appLogger)

	taskHandler := handlers.NewTaskHandler(createTaskCmd, getTaskQuery, appLogger)

	// 6. Setup Server
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
