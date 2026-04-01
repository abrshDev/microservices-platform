# User Service

A backend microservice built with Go, implementing a dual-protocol interface for both REST and gRPC communication. This project follows the CQRS (Command Query Responsibility Segregation) pattern and is designed to serve as the identity provider for a multi-service architecture.

## Project Structure

user\_service/
├── cmd/
│   └── main.go
├── internal/
│   ├── app/
│   │   └── user/
│   │       ├── commands/
│   │       │   ├── create\_user.go
│   │       │   └── delete_user.go
│   │       └── queries/
│   │           ├── get_user.go
│   │           └── login_user.go
│   ├── delivery/
│   │   └── http/
│   │       ├── handlers/
│   │       │   └── user_handler.go
│   │       ├── middleware/
│   │       │   └── auth.go
│   │       └── router.go
│   ├── domain/
│   │   ├── entities/
│   │   │   └── user.go
│   │   ├── errors/
│   │   │   └── errors.go
│   │   └── repositories/
│   │       └── user_repository.go
│   ├── infrastructure/
│   │   ├── config/
│   │   │   └── config.go
│   │   └── database/
│   │       └── postgres/
│   │           ├── connection.go
│   │           └── user_repository.go
│   └── transport/
│       └── grpc/
│           ├── handlers/
│           │   └── user_grpc_handler.go
│           └── proto/
│               ├── user.pb.go
│               └── user_grpc.pb.go
├── schema/
│   └── grpc/
│       └── user.proto
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum

## Architecture Overview

The project is organized into layers to ensure a clean separation of concerns:

  - cmd: Application entry point and server initialization.
  - internal/app: Core business logic divided into Commands (write) and Queries (read).
  - internal/domain: Entities, repository interfaces, and domain-specific errors.
  - internal/infrastructure: External concerns like database connections and GORM repository implementations.
  - internal/delivery/http: REST API implementation using the Fiber framework.
  - internal/transport/grpc: gRPC server implementation and generated Protocol Buffer code.
  - schema/grpc: Source .proto files defining the service contract.

## Tech Stack

  - Language: Go (Golang)
  - Web Framework: Fiber (REST)
  - Communication: gRPC (HTTP/2 + Protocol Buffers)
  - Database: PostgreSQL
  - ORM: GORM
  - Containerization: Docker and Docker Compose

## Getting Started

### Prerequisites

  - Go 1.21 or higher
  - Docker and Docker Compose
  - Protocol Buffer Compiler (protoc)

### Installation

1.  Clone the repository:
    git clone [https://github.com/abrshDev/user-service.git](https://www.google.com/search?q=https://github.com/abrshDev/user-service.git)
    cd user-service

2.  Synchronize dependencies:
    go mod tidy
    go mod vendor

### Running the Service

The easiest way to run the service along with the database is using Docker Compose:

docker-compose up --build

The service will be available at:

  - REST API: http://localhost:8080
  - gRPC Server: localhost:50051

## API Endpoints

### REST (HTTP)

  - POST /users: Create a new user
  - GET /users/:id: Retrieve user by ID
  - POST /login: Authenticate user
  - DELETE /users/:id: Remove a user

### gRPC (Internal)

  - UserService/GetUser: Optimized binary endpoint for inter-service communication.

## Development

### Generating gRPC Code

If you modify schema/grpc/user.proto, regenerate the Go code using:

protoc --go_out=. --go_opt=paths=source_relative  
--go-grpc_out=. --go-grpc_opt=paths=source_relative  
internal/transport/grpc/proto/user.proto

