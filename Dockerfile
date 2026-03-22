# Step 1: Build the binary
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
# We'll create cmd/main.go soon
RUN go build -o main ./cmd/main.go

# Step 2: Run the binary
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080 50051
CMD ["./main"]