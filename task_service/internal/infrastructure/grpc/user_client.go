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
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &user.GetUserRequest{Id: userID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	return resp, nil
}
func (c *UserClient) Close() error {
	return c.conn.Close()
}
