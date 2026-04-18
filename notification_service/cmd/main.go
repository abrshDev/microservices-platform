package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/abrshDev/notification_service/internal/app/notification/commands"
	"github.com/abrshDev/notification_service/internal/infrastructure/config"
	"github.com/abrshDev/notification_service/internal/infrastructure/database/postgres"
	"github.com/abrshDev/notification_service/internal/infrastructure/kafka" // Ensure this path matches your folder
	"github.com/abrshDev/notification_service/internal/transport/grpc/handlers"
	"github.com/abrshDev/notification_service/internal/transport/grpc/proto/notification"
	g "google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	config.LoadEnv()
	db, err := postgres.NewConnection()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	logger.Info("Database connection and migrations successful")
	repo := postgres.NewEventRepository(db)
	sendHandler := commands.NewSendNotificationHandler(logger, repo)
	// 1. Initialize Infrastructure (Kafka Consumer)
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:29092"
	}

	consumer := kafka.NewNotificationConsumer(
		[]string{kafkaBrokers},
		"task-events",
		"notification-group",
		logger,
		sendHandler,
	)

	// 2. Start Kafka Consumer in a Goroutine (Non-blocking)
	ctx := context.Background()
	go func() {
		logger.Info("Starting Kafka Consumer background worker")
		consumer.Start(ctx)
	}()

	// 3. Initialize CQRS Handler (For gRPC if still needed)

	// 4. Initialize gRPC Transport Handler
	grpcHandler := handlers.NewNotificationGRPCHandler(sendHandler)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := g.NewServer()
	notification.RegisterNotificationServiceServer(server, grpcHandler)

	logger.Info("Notification Service starting", "grpc_port", 50052, "kafka_topic", "task-events")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
