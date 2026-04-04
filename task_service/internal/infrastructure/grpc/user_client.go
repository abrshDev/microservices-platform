package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/abrshDev/task-service/internal/transport/grpc/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserClient struct {
	client user.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserClient(address string) (*UserClient, error) {
	// NewClient is non-blocking; it won't fail if User Service is temporarily down
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &UserClient{
		client: user.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *UserClient) GetUser(ctx context.Context, userID string) (*user.UserResponse, error) {
	var lastErr error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		// 1. Create a per-attempt timeout (Resilience Rule #1: Never wait forever)
		attemptCtx, cancel := context.WithTimeout(ctx, 1*time.Second)

		resp, err := c.client.GetUser(attemptCtx, &user.GetUserRequest{Id: userID})
		cancel()

		if err == nil {
			return resp, nil // Success!
		}

		// 2. Analyze the error (Resilience Rule #2: Only retry what makes sense)
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, nil // User doesn't exist, no point in retrying
			case codes.DeadlineExceeded, codes.Unavailable, codes.ResourceExhausted:
				// These are "Transient" errors - the service might come back!
				lastErr = err
				// 3. Exponential Backoff (Wait: 100ms, 200ms, 400ms)
				waitTime := time.Duration(100*(1<<i)) * time.Millisecond
				time.Sleep(waitTime)
				continue
			}
		}

		// If it's a critical error we can't fix by retrying, return immediately
		return nil, err
	}

	return nil, fmt.Errorf("user service unreachable after %d attempts: %w", maxRetries, lastErr)
}
func (c *UserClient) Close() error {
	return c.conn.Close()
}
