# Task Service

A Go-based microservice responsible for task orchestration within a multi-service architecture. This service manages task lifecycles, validates user ownership via gRPC calls to the User Service, and dispatches asynchronous updates to the Notification Service.

## Project Structure

task_service/
├── cmd/
│   └── main.go
├── internal/
│   ├── app/
│   │   └── task/
│   │       ├── commands/
│   │       │   └── create_task.go
│   │       └── queries/
│   │           └── get_tasks.go
│   ├── delivery/
│   │   └── http/
│   │       ├── handlers/
│   │       │   └── task_handler.go
│   │       └── router.go
│   ├── domain/
│   │   ├── entities/
│   │   │   └── task.go
│   │   └── repositories/
│   │       └── task_repository.go
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── postgres/
│   │   │       ├── connection.go
│   │   │       └── task_repository.go
│   │   └── grpc/
│   │       ├── user_client.go
│   │       └── notification_client.go
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum

## Architecture Overview

The Task Service acts as the central orchestrator for task-related business logic, following a layered architecture:

  - cmd: Application initialization and dependency injection.
  - internal/app: Business logic layer implementing CQRS commands for task operations.
  - internal/delivery/http: REST interface using the Fiber framework for external clients.
  - internal/infrastructure/database: Persistence layer for PostgreSQL using the repository pattern.
  - internal/infrastructure/grpc: Client implementations that facilitate synchronous communication with the User Service and asynchronous triggers for the Notification Service.

## Tech Stack

  - Language: Go (Golang)
  - Web Framework: Fiber (REST)
  - Communication: gRPC (Client-side for inter-service calls)
  - Database: PostgreSQL
  - Containerization: Docker and Docker Compose

## Getting Started

### Prerequisites

  - Go 1.21 or higher
  - Docker and Docker Compose
  - Root-level .env file containing service addresses

### Installation

1.  Navigate to the project directory:
    cd task_service

2.  Synchronize dependencies:
    go mod tidy

### Running the Service

The Task Service is designed to run within the microservices-platform network. Use the root-level Docker Compose to start the full stack:

docker-compose up --build

The service will be available at:

  - REST API: http://localhost:8081

## API Endpoints

### REST (HTTP)

  - POST /api/v1/tasks/create: Initialize a new task (triggers user validation and notification dispatch)
  - GET /api/v1/tasks/:id: Retrieve task details

## Service Dependencies

This service relies on the availability of the following gRPC endpoints:

  - User Service: localhost:50051 (Sync validation)
  - Notification Service: localhost:50052 (Async dispatch)

---

### **The Recap Checklist**
Since you mentioned wanting a recap later, this README now serves as your "Source of Truth." It documents:
1.  **The Dual Nature:** It shows the service is an HTTP server (to you) but a gRPC client (to other services).
2.  **The Infrastructure:** It lists the exact ports needed for the platform to function.
3.  **The Pattern:** It confirms the use of CQRS in the directory structure.

