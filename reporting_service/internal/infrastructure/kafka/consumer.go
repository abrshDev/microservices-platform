package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/repositories"
	"github.com/abrshDev/reporting-service/internal/infrastructure/metrics"
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

	messages := make(chan messageEnvelope)
	var wg sync.WaitGroup

	go func() {
		defer close(messages)
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					logger.Info("Kafka Task Consumer stopped",
						slog.String("topic", topic),
						slog.String("group", groupID),
					)
					return
				}
				logger.Error("failed to read message", slog.String("error", err.Error()))
				continue
			}

			select {
			case messages <- messageEnvelope{msg: m, err: err}:
			case <-ctx.Done():
				logger.Info("Kafka Task Consumer stopped",
					slog.String("topic", topic),
					slog.String("group", groupID),
				)
				return
			}
		}
	}()

	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for envelope := range messages {
				func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error("worker panicked",
								slog.Any("panic", r),
								slog.Int("worker", workerID),
							)
							metrics.EventsErrors.WithLabelValues("panic").Inc()
						}
					}()

					var event TaskEvent
					if err := json.Unmarshal(envelope.msg.Value, &event); err != nil {
						logger.Error("failed to unmarshal event", slog.String("error", err.Error()))
						metrics.EventsErrors.WithLabelValues("unmarshal_failed").Inc()
						return
					}

					var isNew bool
					err := retryWithBackoff(logger, "InsertIfNotExist", 3, func() error {
						var err error
						isNew, err = repo.InsertIfNotExist(ctx, event.ID)
						return err
					})
					if err != nil {
						logger.Error("idempotency check failed after retries",
							slog.String("event_id", event.ID),
							slog.String("error", err.Error()),
						)
						metrics.EventsErrors.WithLabelValues("idempotency_failed").Inc()
						return
					}
					if !isNew {
						logger.Info("duplicate event detected, skipping",
							slog.String("event_id", event.ID),
						)
						metrics.EventsDuplicates.Inc()
						return
					}

					change := 0
					if event.Action == "TASK_CREATED" {
						change = 1
					} else if event.Action == "TASK_DELETED" {
						change = -1
					}

					if change != 0 {
						err = retryWithBackoff(logger, "UpdateWithAudit", 3, func() error {
							return repo.UpdateWithAudit(ctx, event.UserID, event.TenantID, change, event.Action)
						})
						if err != nil {
							logger.Error("audit update failed after retries",
								slog.String("user_id", event.UserID),
								slog.String("error", err.Error()),
							)
							metrics.EventsErrors.WithLabelValues("audit_failed").Inc()
							return
						}
					}

					err = retryWithBackoff(logger, "UpdateStatus", 3, func() error {
						return repo.UpdateStatus(ctx, event.ID, "COMPLETED")
					})
					if err != nil {
						logger.Error("status update failed after retries",
							slog.String("event_id", event.ID),
							slog.String("error", err.Error()),
						)
						metrics.EventsErrors.WithLabelValues("status_failed").Inc()
						return
					}

					metrics.EventsProcessed.Inc()
					logger.Info("event processed",
						slog.String("action", event.Action),
						slog.String("correlation_id", event.CorrelationID),
						slog.String("task_id", event.ID),
						slog.String("user_id", event.UserID),
						slog.Int("worker", workerID),
					)
				}()
			}
		}(w)
	}

	wg.Wait()
}

func retryWithBackoff(logger *slog.Logger, operation string, maxRetries int, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		logger.Warn("operation failed, retrying",
			slog.String("operation", operation),
			slog.Int("attempt", attempt),
			slog.Int("max_retries", maxRetries),
			slog.String("error", lastErr.Error()),
		)
		time.Sleep(time.Duration(attempt*100) * time.Millisecond)
	}
	return fmt.Errorf("%s failed after %d retries: %w", operation, maxRetries, lastErr)
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

	messages := make(chan messageEnvelope)
	var wg sync.WaitGroup

	go func() {
		defer close(messages)
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					logger.Info("Kafka User Consumer stopped",
						slog.String("topic", topic),
						slog.String("group", groupID),
					)
					return
				}
				logger.Error("failed to read user message", slog.String("error", err.Error()))
				continue
			}

			select {
			case messages <- messageEnvelope{msg: m, err: err}:
			case <-ctx.Done():
				logger.Info("Kafka User Consumer stopped",
					slog.String("topic", topic),
					slog.String("group", groupID),
				)
				return
			}
		}
	}()

	for w := 1; w <= 3; w++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for envelope := range messages {
				func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error("worker panicked",
								slog.Any("panic", r),
								slog.Int("worker", workerID),
							)
							metrics.EventsErrors.WithLabelValues("panic").Inc()
						}
					}()

					var event UserEvent
					if err := json.Unmarshal(envelope.msg.Value, &event); err != nil {
						logger.Error("failed to unmarshal user event", slog.String("error", err.Error()))
						metrics.EventsErrors.WithLabelValues("unmarshal_failed").Inc()
						return
					}

					var isNew bool
					err := retryWithBackoff(logger, "InsertIfNotExist", 3, func() error {
						var err error
						isNew, err = repo.InsertIfNotExist(ctx, event.UserID)
						return err
					})
					if err != nil {
						logger.Error("idempotency check failed after retries",
							slog.String("user_id", event.UserID),
							slog.String("error", err.Error()),
						)
						metrics.EventsErrors.WithLabelValues("idempotency_failed").Inc()
						return
					}
					if !isNew {
						logger.Info("duplicate user event detected, skipping",
							slog.String("user_id", event.UserID),
						)
						metrics.EventsDuplicates.Inc()
						return
					}

					err = retryWithBackoff(logger, "UpdateWithAudit", 3, func() error {
						return repo.UpdateWithAudit(ctx, event.UserID, event.TenantID, 0, "USER_REGISTERED")
					})
					if err != nil {
						logger.Error("failed to initialize user summary after retries",
							slog.String("user_id", event.UserID),
							slog.String("error", err.Error()),
						)
						metrics.EventsErrors.WithLabelValues("audit_failed").Inc()
						return
					}

					err = retryWithBackoff(logger, "UpdateStatus", 3, func() error {
						return repo.UpdateStatus(ctx, event.UserID, "COMPLETED")
					})
					if err != nil {
						logger.Error("failed to update status after retries",
							slog.String("user_id", event.UserID),
							slog.String("error", err.Error()),
						)
						metrics.EventsErrors.WithLabelValues("status_failed").Inc()
						return
					}

					metrics.EventsProcessed.Inc()
					logger.Info("user record initialized",
						slog.String("user_id", event.UserID),
						slog.Int("tenant_id", int(event.TenantID)),
						slog.Int("worker", workerID),
					)
				}()
			}
		}(w)
	}

	wg.Wait()
}
