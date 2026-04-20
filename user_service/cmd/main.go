package main

import (
	"log"
	"net"
	"os"

	"github.com/abrshDev/user-service/internal/app/user/commands"
	"github.com/abrshDev/user-service/internal/app/user/queries"
	"github.com/abrshDev/user-service/internal/delivery/http"
	http_handlers "github.com/abrshDev/user-service/internal/delivery/http/handlers"
	"github.com/abrshDev/user-service/internal/infrastructure/config"
	"github.com/abrshDev/user-service/internal/infrastructure/database/postgres"
	"github.com/abrshDev/user-service/internal/infrastructure/kafka" // Kafka infrastructure

	// gRPC Imports
	grpc_handlers "github.com/abrshDev/user-service/internal/transport/grpc/handlers"
	pb "github.com/abrshDev/user-service/internal/transport/grpc/proto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

func main() {
	// 1. Load configuration and connect to Postgres
	config.LoadEnv()
	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 2. Setup Kafka Producer for user events
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:29092"
	}
	userProducer := kafka.NewUserProducer([]string{kafkaBrokers})

	// 3. Initialize Repositories
	userRepo := postgres.NewUserRepository(db)

	// 4. Initialize CQRS Handlers (Injecting the producer into CreateUser)
	createUserCmd := commands.NewCreateUserHandler(userRepo, userProducer)
	getUserQuery := queries.NewGetUserHandler(userRepo)
	deleteUserCmd := commands.NewDeleteUserHandler(userRepo)
	loginQuery := queries.NewLoginHandler(userRepo)
	checkStatusQuery := queries.NewCheckUserStatusHandler(userRepo)

	// 5. Setup gRPC Server
	userGrpcHandler := grpc_handlers.NewUserGRPCHandler(getUserQuery, deleteUserCmd, checkStatusQuery)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen for gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userGrpcHandler)

	go func() {
		log.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// 6. Setup REST API (Fiber)
	userHttpHandler := http_handlers.NewUserHandler(createUserCmd, getUserQuery, deleteUserCmd, loginQuery)
	app := fiber.New()
	http.SetupRoutes(app, userHttpHandler)

	log.Println("REST server listening on :8080")
	log.Fatal(app.Listen(":8080"))
}
