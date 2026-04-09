# Stage 1: Build the binary
# Using 1.25 as requested to match your environment
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
# Pointing to the main.go inside the notification_service structure
RUN CGO_ENABLED=0 GOOS=linux go build -o notification-service ./cmd/main.go

# Stage 2: Final lightweight image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
# Copy the binary from the builder stage
COPY --from=builder /app/notification-service .

# Standardize the gRPC port
EXPOSE 50052

CMD ["./notification-service"]