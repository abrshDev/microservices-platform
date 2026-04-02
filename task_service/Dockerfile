# Stage 1: Build the binary
# Match the Go version in your go.mod (1.25)
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# FIX: Path updated to match your tree (cmd/main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -o task-service ./cmd/main.go

# Stage 2: Final lightweight image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/task-service .

EXPOSE 8081

CMD ["./task-service"]