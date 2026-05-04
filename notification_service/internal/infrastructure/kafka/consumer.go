package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/abrshDev/notification_service/internal/app/notification/commands"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

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

			c.logger.Error("failed to unmarshal message",
				slog.String("raw", string(msg.Value)),
				slog.String("error", err.Error()),
			)
			continue
		}

		cmd := commands.SendNotificationCommand{
			TaskID:  event.TaskID.String(),
			UserID:  event.UserID.String(),
			Message: "New Task Created: " + event.Title,
			Type:    "TASK_ALERT",
		}

		maxRetries := 3
		var lastErr error
		success := false

		for attempt := 1; attempt <= maxRetries; attempt++ {
			lastErr = c.tryExecute(ctx, cmd)
			if lastErr == nil {
				success = true
				break
			}
			c.logger.Warn("handler attempt failed",
				slog.Int("attempt", attempt),
				slog.Int("max_retries", maxRetries),
				slog.String("error", lastErr.Error()),
			)
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}

		if !success {
			c.logger.Error("message exhausted all retries - DEAD LETTER",
				slog.String("payload", string(msg.Value)),
				slog.String("final_error", lastErr.Error()),
			)
		}
	}
}

func (c *NotificationConsumer) tryExecute(ctx context.Context, cmd commands.SendNotificationCommand) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			c.logger.Error("handler panicked",
				slog.Any("panic", r),
				slog.String("task_id", cmd.TaskID),
			)
		}
	}()
	return c.sendHandler.Execute(ctx, cmd)
}
func (c *NotificationConsumer) Close() error {
	return c.reader.Close()
}
