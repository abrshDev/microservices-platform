package main

import (
	"log"
	"net"

	"github.com/abrshDev/user-service/internal/app/user/commands"
	"github.com/abrshDev/user-service/internal/app/user/queries"
	"github.com/abrshDev/user-service/internal/delivery/http"
	http_handlers "github.com/abrshDev/user-service/internal/delivery/http/handlers"
	"github.com/abrshDev/user-service/internal/infrastructure/database/postgres"

	// gRPC Imports
	grpc_handlers "github.com/abrshDev/user-service/internal/transport/grpc/handlers"
	pb "github.com/abrshDev/user-service/internal/transport/grpc/proto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

func main() {
	// 1. Database Connection
	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 2. Repository Initialization
	userRepo := postgres.NewUserRepository(db)

	// 3. CQRS Handlers (Commands & Queries)
	createUserCmd := commands.NewCreateUserHandler(userRepo)
	getUserQuery := queries.NewGetUserHandler(userRepo)
	deleteUserCmd := commands.NewDeleteUserHandler(userRepo)
	loginQuery := queries.NewLoginHandler(userRepo)
	checkStatusQuery := queries.NewCheckUserStatusHandler(userRepo)

	// gRPC server

	// Initialize gRPC Handler
	userGrpcHandler := grpc_handlers.NewUserGRPCHandler(getUserQuery, deleteUserCmd, checkStatusQuery)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen for gRPC: %v", err)
	}

	// Create gRPC Server and Register the Service
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userGrpcHandler)

	// Start gRPC
	go func() {
		log.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// REST API SETUP

	userHttpHandler := http_handlers.NewUserHandler(createUserCmd, getUserQuery, deleteUserCmd, loginQuery)

	app := fiber.New()

	http.SetupRoutes(app, userHttpHandler)

	// Start Fiber (Blocking call)
	log.Println("REST server listening on :8080")
	log.Fatal(app.Listen(":8080"))
}
