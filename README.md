# Microservices Platform

This repository is not a loose collection of services sharing a Docker network. It is a boundary-first Go system where each service owns a decision, a database, and a failure mode.

The platform is built around four business services.

User Service handles identity, login, user lifecycle, and tenant-aware user data.

Task Service owns task creation and task retrieval, validates users over gRPC, and emits domain events when new work is created.

Notification Service listens to task events and turns them into persisted notifications through an asynchronous Kafka consumer.

Reporting Service maintains a tenant-aware read model fed by events, giving the system a separate projection layer instead of forcing every query back through write services.

Taken together, the design already shows the right instincts: synchronous boundaries only where validation must be immediate, asynchronous boundaries where side effects should not hold the request path hostage, separate databases per service, startup health checks, and container restart policies.

# Service Map

| Service | Main Interface | Port | Data Store | Role |
|---|---|---:|---|---|
| user_service | REST and gRPC | 8080, 50051 | PostgreSQL | Identity, authentication, user lifecycle |
| task_service | REST | 8081 | PostgreSQL | Task creation, task lookup, user validation, task events |
| notification_service | gRPC and Kafka consumer | 50052 | PostgreSQL | Event-driven notifications |
| reporting_service | REST | 8083 | PostgreSQL | Tenant-aware reporting projection |
| kafka | Event backbone | 9092 | N/A | Async transport for user and task events |
| zookeeper | Kafka coordination | 2181 | N/A | Broker coordination |

# System Shape

```text
Client
  |
  +--> User Service --------------------------> user_db
  |        |
  |        +--> topic: user-events -----------------------------+
  |                                                             |
  +--> Task Service --------------------------> task_db         |
           |                                                    |
           +--> gRPC call to User Service                       |
           |                                                    |
           +--> topic: task-events -----------+-----------------+
                                               |
                                               +--> Notification Service --> notification_db
                                               |
                                               +--> Reporting Service -----> reporting_db

Reporting Service also consumes user-events to initialize tenant-aware summary rows.
```

# Why The Architecture Works

The important choice in this codebase is not that it uses microservices. The important choice is that it draws sharp lines.

User data stays in the user domain.

Task creation can ask the user domain a direct question over gRPC when immediate consistency matters.

Notifications do not sit in the hot request path. They trail the system through Kafka, which means the user-facing workflow does not need to wait for message formatting, delivery logic, or notification persistence.

Reporting is treated as a projection, not a transactional authority. That is the right move. It keeps read models fast, cheap to evolve, and isolated from write-side contention.

Tenant data is also becoming a first-class concern across the platform. That is a serious architectural step because it pushes the system beyond single-tenant assumptions and forces every boundary to be explicit about ownership and scope.

# What Is Already Strong

1. Each service owns its own PostgreSQL database.
2. Startup is guarded by health checks instead of blind ordering.
3. Kafka decouples side effects from the request path.
4. Task production uses `RequireAll` acknowledgements, which is a good durability choice for domain events.
5. The task service includes graceful shutdown behavior for HTTP, Kafka, gRPC, and the database handle.
6. Reporting is not querying operational services directly for every read. It builds and serves its own summary model.
7. The platform now carries tenant context through user, task, and reporting flows.

# Current Operational Posture

| Concern | Current Posture In This Repository | Why It Matters |
|---|---|---|
| Service ownership | Separate services and separate databases | Limits blast radius and keeps domains honest |
| Startup discipline | Compose waits for healthy databases and Kafka | Avoids race conditions during boot |
| Failure isolation | Notifications and reporting are fed asynchronously | Slow consumers do not have to block writes |
| Event durability | Task producer uses stronger broker acknowledgement settings | Lowers the chance of silent event loss |
| Read scaling | Reporting uses a dedicated projection model | Prevents analytical reads from distorting write services |
| Multi-tenancy | Tenant id flows through user, task, and reporting paths | Makes data boundaries explicit |
| Health visibility | Compose health checks plus a reporting health endpoint | Gives the platform a basic runtime pulse |

# Observability, Resilience, And The Next Hardening Layer

This codebase has the right skeleton. The next leap is to make failures easier to see, easier to contain, and easier to recover from under pressure.

The immediate engineering agenda should look like this.

1. Add request correlation ids across REST, gRPC, and Kafka boundaries.
2. Standardize structured logging across every service, not just parts of the system.
3. Export metrics for request latency, Kafka lag, consumer failures, gRPC error rates, and database pool pressure.
4. Introduce readiness and liveness endpoints in every service, not just reporting and container-level checks.
5. Put circuit breakers and bounded retries around synchronous gRPC calls from task service to user service.
6. Add dead-letter handling and replay strategy for poison Kafka messages.
7. Introduce idempotent consumers or an outbox pattern where cross-service delivery guarantees need to be stronger.

Circuit breakers matter here because your architecture already mixes synchronous validation with asynchronous propagation. That is a healthy design, but only if the synchronous edge is treated with respect. The task service should be able to degrade gracefully when user validation becomes slow, unavailable, or noisy. A circuit breaker there would prevent cascading latency and protect the rest of the system from a sick dependency.

Observability matters just as much. A distributed system becomes expensive the moment it goes dark. Good logs help. Correlated logs, metrics, traces, and explicit health semantics turn a midnight outage into an engineering problem instead of a guessing contest.

# API Surface

| Service | Endpoint | Method | Purpose |
|---|---|---|---|
| user_service | `/api/v1/users` | POST | Create a user |
| user_service | `/api/v1/login` | POST | Authenticate and issue access flow |
| user_service | `/api/v1/users/:id` | GET | Fetch a user by id |
| user_service | `/api/v1/users/:id` | DELETE | Remove a user |
| task_service | `/api/v1/tasks/create` | POST | Create a task |
| task_service | `/api/v1/tasks/get/:id` | GET | Fetch a task |
| reporting_service | `/api/v1/reports/:tenantId/:userId` | GET | Read tenant-aware task summary |
| reporting_service | `/health` | GET | Service health check |

Internal traffic also uses gRPC and Kafka.

User Service exposes gRPC on port `50051`.

Notification Service exposes gRPC on port `50052`.

Kafka topics currently in use include `user-events` and `task-events`.

# Running The Platform

## Prerequisites

1. Docker
2. Docker Compose
3. A root `.env` file with the database credentials, service addresses, Kafka broker address, and JWT secret expected by the services

## Start

```bash
docker compose up --build
```

## Default Local Ports

| Component | Port |
|---|---:|
| User REST | 8080 |
| User gRPC | 50051 |
| Task REST | 8081 |
| Notification gRPC | 50052 |
| Reporting REST | 8083 |
| Kafka external listener | 9092 |
| User Postgres | 5432 |
| Task Postgres | 5433 |
| Notification Postgres | 5436 |
| Reporting Postgres | 5437 |

# Repository Layout

```text
microservices-platform/
  docker-compose.yml
  user_service/
  task_service/
  notification_service/
  reporting_service/
```

Each service follows a layered shape centered around `cmd`, `internal`, and protocol or migration assets. The codebase is moving toward a model where commands, queries, transport adapters, infrastructure, and persistence remain visible as separate concerns. That separation is worth protecting. It will make the next phase of hardening much easier.

# Engineering Notes

This platform is already more interesting than a standard CRUD demo because it is trying to solve the real distributed-systems problem: how to preserve clear ownership while still letting information move across the system at the right time and with the right guarantees.

That is the right ambition.

The next milestone is not more endpoints. It is operational maturity.

When this repository gains end-to-end tracing, circuit breakers on synchronous edges, replay-safe event handling, stronger health semantics, and measurable service-level objectives, it will stop looking like a promising microservices exercise and start looking like the foundation of a production platform.

That is a very reachable next step from where the code stands today.
