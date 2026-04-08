package main

import (
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/abrshDev/notification_service/internal/app/notification/commands"
	"github.com/abrshDev/notification_service/internal/transport/grpc/handlers"
	"github.com/abrshDev/notification_service/internal/transport/grpc/proto/notification"
	g "google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Initialize CQRS Handler
	sendHandler := commands.NewSendNotificationHandler(logger)

	// Initialize Transport Handler
	grpcHandler := handlers.NewNotificationGRPCHandler(sendHandler)

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := g.NewServer()
	notification.RegisterNotificationServiceServer(server, grpcHandler)

	logger.Info("Notification Service (CQRS) starting", "port", 8082)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
