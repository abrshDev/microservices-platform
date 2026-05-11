package main

import (
	"context"
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
	"github.com/abrshDev/task-service/internal/infrastructure/outbox"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config.LoadEnv()

	appLogger := logger.NewLogger("task-service")

	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:29092"
	}

	taskProducer := kafka.NewEventProducer([]string{kafkaBrokers}, "task-events")

	userSvcAddr := os.Getenv("USER_SERVICE_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "user-service:50051"
	}

	userClient, err := grpc.NewUserClient(userSvcAddr)
	if err != nil {
		log.Fatalf("Failed to initialize User gRPC client: %v", err)
	}

	taskRepo := postgres.NewTaskRepository(db)
	outboxRepo := postgres.NewOutboxRepository(db)

	createTaskCmd := commands.NewCreateTaskHandler(taskRepo, userClient, taskProducer, appLogger)
	getTaskQuery := queries.NewGetTaskHandler(taskRepo, userClient, appLogger)
	deleteTaskcmd := commands.NewDeleteTaskHandler(taskRepo, userClient, appLogger)
	taskHandler := handlers.NewTaskHandler(createTaskCmd, getTaskQuery, deleteTaskcmd, appLogger)

	// Outbox worker
	outboxWorker := outbox.NewWorker(outboxRepo, taskProducer, appLogger)

	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go outboxWorker.Start(ctx)

	// HTTP server
	app := fiber.New()
	http.SetupRoutes(app, taskHandler)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		port := os.Getenv("TASK_APP_PORT")
		if port == "" {
			port = "8081"
		}
		log.Printf("Task Service running on %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Fiber stopped: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down...")

	cancel()

	if err := app.Shutdown(); err != nil {
		log.Printf("Fiber shutdown failed: %v", err)
	}
	if err := taskProducer.Close(); err != nil {
		log.Printf("Kafka close failed: %v", err)
	}
	userClient.Close()

	sqlDB, _ := db.DB()
	sqlDB.Close()

	log.Println("Exited cleanly")
}
