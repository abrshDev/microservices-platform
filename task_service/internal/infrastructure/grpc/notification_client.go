package grpc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/abrshDev/task-service/internal/transport/grpc/proto/notification"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationClient struct {
	client notification.NotificationServiceClient
	logger *slog.Logger
}

func NewNotificationClient(target string, logger *slog.Logger) (*NotificationClient, error) {
	// Dial the notification service
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service at %s: %w", target, err)
	}

	return &NotificationClient{
		client: notification.NewNotificationServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *NotificationClient) SendTaskNotification(ctx context.Context, userID, taskTitle string) {
	_, err := c.client.SendNotification(ctx, &notification.NotificationRequest{
		UserId:  userID,
		Message: "Success! New task created: " + taskTitle,
		Type:    "TASK_ALERT",
	})

	if err != nil {
		// We log the error but don't return it because we don't want
		// to fail the task creation just because the notification failed.
		c.logger.Error("failed to send gRPC notification", "error", err)
	}
}
