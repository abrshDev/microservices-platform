# Microservices Platform

This repository is a Go microservices platform built around clear service ownership, separate databases, gRPC for synchronous validation, and Kafka for asynchronous propagation.

The current stack includes:

- `user_service` for identity, login, user lifecycle, and tenant-aware user records
- `task_service` for task creation, retrieval, and deletion with synchronous user validation
- `notification_service` for asynchronous notification handling from Kafka
- `reporting_service` for tenant-aware read models and audit history

## What Was Finished Since The Last Pass

The README previously listed several hardening items as future work. The following pieces are now implemented in code:

- Correlation IDs are generated or accepted by the task service with `X-Correlation-ID`, attached to outbound gRPC metadata, added to task events, and logged by the reporting consumer.
- Task service now uses structured JSON logging, timeout-bound retries, and a circuit breaker around calls to the user service.
- Task creation uses a transactional outbox. The task row and outbox event are written together, and a background worker publishes pending events to Kafka.
- Reporting and notification consumers are idempotent through `processed_events` tables.
- Reporting includes retry-with-backoff, worker fan-out, audit logging, and Prometheus metrics exposed at `/metrics`.
- Tenant context now flows through user creation, task persistence, task events, and reporting summaries.
- Task HTTP routes are rate-limited to `5` requests per minute per IP.

## Service Map

| Service | Interfaces | Ports | Data Store | Role |
|---|---|---|---|---|
| `user_service` | REST, gRPC, Kafka producer | internal `8080`, internal `50051` | PostgreSQL | Identity, login, user lifecycle |
| `task_service` | REST, gRPC client, Kafka outbox producer | internal `8081` | PostgreSQL | Task create/get/delete, user validation |
| `notification_service` | Kafka consumer, gRPC server | internal `50052` | PostgreSQL | Idempotent notification processing |
| `reporting_service` | REST, Kafka consumers | internal `8083` | PostgreSQL | Reporting projection, audit trail, metrics |
| `traefik` | HTTP gateway | host `90`, host `8080` dashboard | N/A | Routes REST traffic to services |
| `kafka` | Event backbone | host `9092`, internal `29092` | N/A | Async transport for `user-events` and `task-events` |
| `zookeeper` | Kafka coordination | internal `2181` | N/A | Broker coordination |

## System Shape

```text
Client
  |
  +--> Traefik (localhost:90)
          |
          +--> User Service ------------------------> user_db
          |        |
          |        +--> topic: user-events ----------------------+
          |                                                      |
          +--> Task Service ------------------------> task_db     |
                   |                                              |
                   +--> gRPC call to User Service                 |
                   |                                              |
                   +--> outbox_events table                       |
                   |       |
                   |       +--> Outbox Worker --> topic: task-events ---+
                   |                                                     |
                   +-----------------------------------------------------+
                                                                         |
                                +--> Notification Service --> notification_db
                                |
                                +--> Reporting Service -----> reporting_db
                                         |
                                         +--> audit_logs
                                         +--> processed_events
                                         +--> /metrics
```

## Runtime Behavior

- `user_service` creates users, hashes passwords, issues JWTs on login, and publishes `user-events` with tenant context.
- `task_service` validates user status and role over gRPC before creating or deleting tasks.
- `task_service` fetches user details over gRPC for enriched task reads and degrades gracefully if that read-side call fails.
- `task-events` are published from the outbox worker with Kafka `RequireAll` acknowledgements.
- `notification_service` consumes `task-events`, retries handler failures, and skips duplicates using `processed_events`.
- `reporting_service` consumes both `user-events` and `task-events`, updates a tenant-aware summary, and records audit rows for changes.

## Implemented Hardening

| Concern | Current Implementation |
|---|---|
| Startup discipline | Compose waits on healthy databases and Kafka before dependent services start |
| Failure isolation | Notification and reporting work stays off the request path through Kafka |
| Correlation | Task requests use `X-Correlation-ID` and propagate it into gRPC and Kafka task events |
| Structured logs | Implemented in task, reporting, and notification services via `slog` JSON handlers |
| Sync dependency protection | Task service wraps user-service gRPC calls with a circuit breaker and retries |
| Event durability | Task service uses a transactional outbox plus Kafka `RequireAll` acknowledgements |
| Consumer safety | Reporting and notification consumers are idempotent via `processed_events` |
| Retry strategy | Reporting and notification consumers retry transient failures with backoff |
| Read model isolation | Reporting serves its own projection instead of querying write services directly |
| Multi-tenancy | Tenant IDs are persisted in user, task, and reporting flows |
| Metrics | Reporting exposes Prometheus counters for processed, duplicate, and failed events |
| Auditability | Reporting stores audit rows whenever task summary totals change |

## Still Open

These items are still not fully implemented across the platform:

- Distributed tracing export is not wired up yet.
- Readiness and liveness endpoints are not available in every service. `reporting_service` currently exposes `/health`.
- Metrics are only exposed by `reporting_service`.
- A real dead-letter topic and replay workflow are not implemented. `notification_service` currently logs exhausted messages as dead-letter candidates.
- `user_service` publishes events directly and does not yet use an outbox pattern.
- `reporting_service` can process `TASK_DELETED`, but `task_service` does not currently emit delete events, so projections only advance from created-task events.

## API Surface

### External HTTP Access

When the stack is started with Docker Compose, REST traffic is routed through Traefik on `http://localhost:90`.

| Service | Endpoint | Method | Notes |
|---|---|---|---|
| `user_service` | `/api/v1/users` | `POST` | Create a user |
| `user_service` | `/api/v1/login` | `POST` | Authenticate and issue JWT |
| `user_service` | `/api/v1/users/:id` | `GET` | Protected by Bearer token |
| `user_service` | `/api/v1/users/:id` | `DELETE` | Protected by Bearer token |
| `task_service` | `/api/v1/tasks/create` | `POST` | Rate-limited; accepts and returns `X-Correlation-ID` |
| `task_service` | `/api/v1/tasks/get/:id` | `GET` | Rate-limited; returns enriched user data when gRPC read succeeds |
| `task_service` | `/api/v1/tasks/delete/:id` | `DELETE` | Rate-limited; requires `X-User-ID` header |
| `reporting_service` | `/api/v1/reports/:tenantId/:userId` | `GET` | Tenant-aware summary lookup |
| `reporting_service` | `/health` | `GET` | Simple health endpoint |
| `reporting_service` | `/metrics` | `GET` | Prometheus metrics |

### Internal-Only Interfaces

- `user_service` gRPC listens on `user-service:50051` inside the Compose network.
- `notification_service` gRPC listens on `notification-service:50052` inside the Compose network.
- Kafka topics in active use are `user-events` and `task-events`.

## Running The Platform

### Prerequisites

1. Docker
2. Docker Compose
3. A root `.env` file containing database credentials, Kafka settings, service addresses, and `JWT_SECRET`

### Start

```bash
docker compose up --build
```

### Host-Exposed Ports

| Component | Host Port |
|---|---:|
| Traefik gateway | 90 |
| Traefik dashboard | 8080 |
| Kafka | 9092 |
| User Postgres | 5666 |
| Task Postgres | 5433 |
| Notification Postgres | 5436 |
| Reporting Postgres | 5437 |

### Internal Container Ports

| Component | Internal Port |
|---|---:|
| User REST | 8080 |
| User gRPC | 50051 |
| Task REST | 8081 |
| Notification gRPC | 50052 |
| Reporting REST | 8083 |
| Kafka internal listener | 29092 |
| Zookeeper | 2181 |

## Repository Layout

```text
microservices-platform/
  docker-compose.yml
  README.md
  user_service/
  task_service/
  notification_service/
  reporting_service/
```

Each service keeps its own `cmd`, `internal`, and schema or migration assets so business logic, transport adapters, persistence, and infrastructure concerns stay separated.
