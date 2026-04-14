package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/abrshDev/task-service/internal/app/task/commands"
	"github.com/abrshDev/task-service/internal/app/task/queries"
	"github.com/abrshDev/task-service/internal/delivery/http"
	"github.com/abrshDev/task-service/internal/delivery/http/handlers"
	"github.com/abrshDev/task-service/internal/infrastructure/config"
	"github.com/abrshDev/task-service/internal/infrastructure/database/postgres"
	"github.com/abrshDev/task-service/internal/infrastructure/grpc"
	"github.com/abrshDev/task-service/internal/infrastructure/kafka"
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
		kafkaBrokers = "kafka:29092"
	}
	taskProducer := kafka.NewEventProducer([]string{kafkaBrokers}, "task-events")

	// 3. Service Discovery Addresses
	userSvcAddr := os.Getenv("USER_SERVICE_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "user-service:50051"
	}

	// 4. Initialize gRPC Client
	userClient, err := grpc.NewUserClient(userSvcAddr)
	if err != nil {
		log.Fatalf("Failed to initialize User gRPC client: %v", err)
	}

	// 5. Dependency Injection [cite: 71, 84]
	taskRepo := postgres.NewTaskRepository(db)
	createTaskCmd := commands.NewCreateTaskHandler(taskRepo, userClient, taskProducer, appLogger)
	getTaskQuery := queries.NewGetTaskHandler(taskRepo, userClient, appLogger)
	taskHandler := handlers.NewTaskHandler(createTaskCmd, getTaskQuery, appLogger)

	// 6. Setup Server
	app := fiber.New()
	http.SetupRoutes(app, taskHandler)

	// --- GRACEFUL SHUTDOWN LOGIC ---
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		port := os.Getenv("TASK_APP_PORT")
		if port == "" {
			port = "8081"
		}
		log.Printf("Task Service is starting on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Fiber server stopped: %v", err)
		}
	}()

	// Wait for the signal
	<-sigChan
	log.Println("\nShutting down Task Service...")

	// 1. Shutdown Fiber
	if err := app.Shutdown(); err != nil {
		log.Printf("Fiber shutdown failed: %v", err)
	}

	// 2. Close Kafka Producer [cite: 81]
	if err := taskProducer.Close(); err != nil {
		log.Printf("Kafka producer close failed: %v", err)
	}

	// 3. Close gRPC Client [cite: 32]
	userClient.Close()

	// 4. Close DB connection [cite: 40]
	sqlDB, _ := db.DB()
	sqlDB.Close()

	log.Println("Task Service exited gracefully.")
}
