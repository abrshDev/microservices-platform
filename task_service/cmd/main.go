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
	// load env vars
	config.LoadEnv()

	// logger
	appLogger := logger.NewLogger("task-service")

	// db connection
	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// kafka setup
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:29092"
	}
	taskProducer := kafka.NewEventProducer([]string{kafkaBrokers}, "task-events")

	// user service address
	userSvcAddr := os.Getenv("USER_SERVICE_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "user-service:50051"
	}

	// grpc client
	userClient, err := grpc.NewUserClient(userSvcAddr)
	if err != nil {
		log.Fatalf("Failed to initialize User gRPC client: %v", err)
	}

	// wire dependencies
	taskRepo := postgres.NewTaskRepository(db)
	createTaskCmd := commands.NewCreateTaskHandler(taskRepo, userClient, taskProducer, appLogger)
	getTaskQuery := queries.NewGetTaskHandler(taskRepo, userClient, appLogger)
	taskHandler := handlers.NewTaskHandler(createTaskCmd, getTaskQuery, appLogger)

	// http server
	app := fiber.New()
	http.SetupRoutes(app, taskHandler)

	// graceful shutdown setup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// run server
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

	// wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")

	// stop http
	if err := app.Shutdown(); err != nil {
		log.Printf("Fiber shutdown failed: %v", err)
	}

	// close kafka
	if err := taskProducer.Close(); err != nil {
		log.Printf("Kafka close failed: %v", err)
	}

	// close grpc
	userClient.Close()

	// close db
	sqlDB, _ := db.DB()
	sqlDB.Close()

	log.Println("Exited cleanly")
}
