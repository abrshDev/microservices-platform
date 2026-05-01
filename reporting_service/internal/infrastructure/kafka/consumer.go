package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/repositories"
	"github.com/segmentio/kafka-go"
)

type TaskEvent struct {
	CorrelationID string    `json:"correlation_id"`
	ID            string    `json:"task_id"`
	UserID        string    `json:"user_id"`
	TenantID      uint64    `json:"tenant_id"`
	Action        string    `json:"action"`
	Timestamp     time.Time `json:"timestamp"`
}

type UserEvent struct {
	UserID    string    `json:"user_id"`
	TenantID  uint64    `json:"tenant_id"`
	Email     string    `json:"email"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

func StartTaskConsumer(brokers []string, topic string, groupID string, repo repositories.SummaryRepo, ctx context.Context, logger *slog.Logger) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer reader.Close()

	logger.Info("Kafka Task Consumer started",
		slog.String("topic", topic),
		slog.String("group", groupID),
	)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Error("failed to read message", slog.String("error", err.Error()))
			continue
		}

		var event TaskEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			logger.Error("failed to unmarshal event", slog.String("error", err.Error()))
			continue
		}

		isNew, err := repo.InsertIfNotExist(ctx, event.ID)
		if err != nil {
			logger.Error("failed to check idempotency",
				slog.String("event_id", event.ID),
				slog.String("error", err.Error()),
			)
			continue
		}
		if !isNew {
			logger.Info("duplicate event detected, skipping",
				slog.String("event_id", event.ID),
				slog.String("correlation_id", event.CorrelationID),
			)
			continue
		}

		change := 0
		if event.Action == "TASK_CREATED" {
			change = 1
		} else if event.Action == "TASK_DELETED" {
			change = -1
		}

		if change != 0 {
			err := repo.UpdateWithAudit(event.UserID, event.TenantID, change, event.Action)
			if err != nil {
				logger.Error("failed to process audit update",
					slog.String("user_id", event.UserID),
					slog.String("action", event.Action),
					slog.String("error", err.Error()),
				)
				continue
			}
		}

		repo.UpdateStatus(ctx, event.ID, "COMPLETED")

		logger.Info("event processed",
			slog.String("action", event.Action),
			slog.String("correlation_id", event.CorrelationID),
			slog.String("task_id", event.ID),
			slog.String("user_id", event.UserID),
		)
	}
}

func StartUserConsumer(brokers []string, topic string, groupID string, repo repositories.SummaryRepo, ctx context.Context, logger *slog.Logger) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	defer reader.Close()

	logger.Info("Kafka User Consumer started",
		slog.String("topic", topic),
		slog.String("group", groupID),
	)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Error("failed to read user message", slog.String("error", err.Error()))
			continue
		}

		var event UserEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			logger.Error("failed to unmarshal user event", slog.String("error", err.Error()))
			continue
		}

		isNew, err := repo.InsertIfNotExist(ctx, event.UserID)
		if err != nil {
			logger.Error("failed to check idempotency",
				slog.String("user_id", event.UserID),
				slog.String("error", err.Error()),
			)
			continue
		}
		if !isNew {
			logger.Info("duplicate user event detected, skipping",
				slog.String("user_id", event.UserID),
			)
			continue
		}

		err = repo.UpdateWithAudit(event.UserID, event.TenantID, 0, "USER_REGISTERED")
		if err != nil {
			logger.Error("failed to initialize user summary",
				slog.String("user_id", event.UserID),
				slog.String("error", err.Error()),
			)
			continue
		}

		repo.UpdateStatus(ctx, event.UserID, "COMPLETED")

		logger.Info("user record initialized",
			slog.String("user_id", event.UserID),
			slog.Int("tenant_id", int(event.TenantID)),
		)
	}
}
