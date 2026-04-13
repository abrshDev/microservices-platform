package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

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
	reader *kafka.Reader
	logger *slog.Logger
}

func NewNotificationConsumer(brokers []string, topic string, groupID string, logger *slog.Logger) *NotificationConsumer {
	return &NotificationConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		logger: logger,
	}
}

func (c *NotificationConsumer) Start(ctx context.Context) {
	c.logger.Info("Notification Kafka consumer started", slog.String("topic", c.reader.Config().Topic))

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			c.logger.Error("failed to read message from kafka", slog.String("error", err.Error()))
			continue
		}

		var event TaskCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal task event", slog.String("error", err.Error()))
			continue
		}

		c.logger.Info("successfully consumed kafka event",
			slog.String("task_id", event.TaskID.String()),
			slog.String("user_id", event.UserID.String()),
		)

		// This replaces your old gRPC logic
		fmt.Printf("\n[NOTIFICATION] Alerting User %s: You have a new task: %s\n", event.UserID, event.Title)
	}
}

func (c *NotificationConsumer) Close() error {
	return c.reader.Close()
}
