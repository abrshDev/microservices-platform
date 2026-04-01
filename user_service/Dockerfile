# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy everything (including your vendor folder)
COPY . .

# Build using the local vendor folder (-mod=vendor)
# No internet required here
RUN go build -mod=vendor -o main ./cmd/main.go

# Final stage
FROM alpine:latest

# We removed 'apk add' to avoid the DNS timeout
WORKDIR /root/

# Copy the binary and .env from builder
COPY --from=builder /app/main .


EXPOSE 8080 50051

CMD ["./main"]