package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/abrshDev/task-service/internal/transport/grpc/proto/user"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserClient struct {
	client  user.UserServiceClient
	conn    *grpc.ClientConn
	breaker *gobreaker.CircuitBreaker
}

func NewUserClient(address string) (*UserClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// 3. Initialize the Circuit Breaker settings
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "user-service-breaker",
		MaxRequests: 3,                // Max requests allowed when "half-open"
		Interval:    5 * time.Second,  // Reset interval
		Timeout:     10 * time.Second, // How long to stay "Open" before trying again
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip if we fail 5 times in a row
			return counts.ConsecutiveFailures > 5
		},
	})

	return &UserClient{
		client:  user.NewUserServiceClient(conn),
		conn:    conn,
		breaker: cb, // 4. Assign it
	}, nil
}

func (c *UserClient) GetUser(ctx context.Context, userID string) (*user.UserResponse, error) {
	// 5. Wrap your logic inside the breaker's Execute function
	result, err := c.breaker.Execute(func() (interface{}, error) {
		var lastErr error
		maxRetries := 3

		for i := 0; i < maxRetries; i++ {
			attemptCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
			resp, err := c.client.GetUser(attemptCtx, &user.GetUserRequest{Id: userID})
			cancel()

			if err == nil {
				return resp, nil
			}

			st, ok := status.FromError(err)
			if ok {
				switch st.Code() {
				case codes.NotFound:
					// IMPORTANT: Don't return error to breaker for 404s,
					// otherwise the breaker will trip because of missing users!
					return nil, nil
				case codes.DeadlineExceeded, codes.Unavailable, codes.ResourceExhausted:
					lastErr = err
					waitTime := time.Duration(100*(1<<uint(i))) * time.Millisecond
					time.Sleep(waitTime)
					continue
				}
			}
			return nil, err
		}
		return nil, lastErr
	})

	if err != nil {
		// This could be a gRPC error OR a "circuit breaker is open" error
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*user.UserResponse), nil
}

func (c *UserClient) Close() error {
	return c.conn.Close()
}
