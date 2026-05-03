package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
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

type messageEnvelope struct {
	msg kafka.Message
	err error
}

/////////////////////////////////////////////////////////
// TASK CONSUMER
/////////////////////////////////////////////////////////

func StartTaskConsumer(
	brokers []string,
	topic string,
	groupID string,
	repo repositories.SummaryRepo,
	ctx context.Context,
	logger *slog.Logger,
) {
	var wg sync.WaitGroup

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	logger.Info("Task Consumer started",
		slog.String("topic", topic),
	)

	messages := make(chan messageEnvelope)

	// Reader goroutine
	go func() {
		defer close(messages)

		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Error("read error", slog.String("error", err.Error()))
				return
			}

			select {
			case messages <- messageEnvelope{msg: m}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Workers
	for w := 1; w <= 3; w++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for envelope := range messages {
				var event TaskEvent

				if err := json.Unmarshal(envelope.msg.Value, &event); err != nil {
					logger.Error("unmarshal failed",
						slog.String("error", err.Error()),
						slog.Int("worker", workerID),
					)
					continue
				}

				isNew, err := repo.InsertIfNotExist(ctx, event.ID)
				if err != nil {
					logger.Error("idempotency failed",
						slog.String("event_id", event.ID),
					)
					continue
				}

				if !isNew {
					continue
				}

				change := 0
				switch event.Action {
				case "TASK_CREATED":
					change = 1
				case "TASK_DELETED":
					change = -1
				}

				if change != 0 {
					if err := repo.UpdateWithAudit(event.UserID, event.TenantID, change, event.Action); err != nil {
						logger.Error("update failed",
							slog.String("error", err.Error()),
						)
						continue
					}
				}

				repo.UpdateStatus(ctx, event.ID, "COMPLETED")

				logger.Info("task processed",
					slog.String("task_id", event.ID),
					slog.Int("worker", workerID),
				)
			}

			logger.Info("task worker exit", slog.Int("worker", workerID))
		}(w)
	}

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("Task consumer shutting down...")

	wg.Wait()
}

/////////////////////////////////////////////////////////
// USER CONSUMER
/////////////////////////////////////////////////////////

func StartUserConsumer(
	brokers []string,
	topic string,
	groupID string,
	repo repositories.SummaryRepo,
	ctx context.Context,
	logger *slog.Logger,
) {
	var wg sync.WaitGroup

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	defer reader.Close()

	logger.Info("User Consumer started",
		slog.String("topic", topic),
	)

	messages := make(chan messageEnvelope)

	// Reader goroutine
	go func() {
		defer close(messages)

		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Error("read error", slog.String("error", err.Error()))
				return
			}

			select {
			case messages <- messageEnvelope{msg: m}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Workers
	for w := 0; w < 3; w++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for envelope := range messages {
				var event UserEvent

				if err := json.Unmarshal(envelope.msg.Value, &event); err != nil {
					logger.Error("unmarshal failed",
						slog.String("error", err.Error()),
						slog.Int("worker", workerID),
					)
					continue
				}

				isNew, err := repo.InsertIfNotExist(ctx, event.UserID)
				if err != nil {
					logger.Error("idempotency failed",
						slog.String("user_id", event.UserID),
					)
					continue
				}

				if !isNew {
					continue
				}

				if err := repo.UpdateWithAudit(event.UserID, event.TenantID, 0, "USER_REGISTERED"); err != nil {
					logger.Error("init user failed",
						slog.String("error", err.Error()),
					)
					continue
				}

				repo.UpdateStatus(ctx, event.UserID, "COMPLETED")

				logger.Info("user initialized",
					slog.String("user_id", event.UserID),
					slog.Int("worker", workerID),
				)
			}

			logger.Info("user worker exit", slog.Int("worker", workerID))
		}(w)
	}

	<-ctx.Done()
	logger.Info("User consumer shutting down...")

	wg.Wait()
}
