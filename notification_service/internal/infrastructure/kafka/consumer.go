package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/abrshDev/notification_service/internal/app/notification/commands"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// We define the struct here so we don't need to import from task-service
type TaskCreatedEvent struct {
	TaskID      uuid.UUID `json:"task_id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type NotificationConsumer struct {
	reader      *kafka.Reader
	logger      *slog.Logger
	sendHandler *commands.SendNotificationHandler
}

func NewNotificationConsumer(brokers []string, topic string, groupID string, logger *slog.Logger, sendHandler *commands.SendNotificationHandler) *NotificationConsumer {
	return &NotificationConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		logger:      logger,
		sendHandler: sendHandler,
	}
}

func (c *NotificationConsumer) Start(ctx context.Context) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			c.logger.Error("failed to read message", slog.String("error", err.Error()))
			continue
		}

		var event TaskCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}

		// Logic Change: Use the Command Handler instead of fmt.Printf
		cmd := commands.SendNotificationCommand{
			UserID:  event.UserID.String(),
			Message: "New Task Created: " + event.Title,
			Type:    "TASK_ALERT",
		}

		if err := c.sendHandler.Handle(ctx, cmd); err != nil {
			c.logger.Error("handler failed", slog.String("error", err.Error()))
		}
	}
}
func (c *NotificationConsumer) Close() error {
	return c.reader.Close()
}
