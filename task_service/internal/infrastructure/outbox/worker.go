package outbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	taskevent "github.com/abrshDev/task-service/internal/domain/events"
	"github.com/abrshDev/task-service/internal/domain/repositories"
	"github.com/abrshDev/task-service/internal/infrastructure/kafka"
)

type Worker struct {
	outboxRepo repositories.OutboxRepository
	producer   *kafka.EventProducer
	logger     *slog.Logger
}

func NewWorker(outboxRepo repositories.OutboxRepository, producer *kafka.EventProducer, logger *slog.Logger) *Worker {
	return &Worker{
		outboxRepo: outboxRepo,
		producer:   producer,
		logger:     logger,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processPendingEvents(ctx)
		case <-ctx.Done():
			w.logger.Info("outbox worker stopped")
			return
		}
	}
}

func (w *Worker) processPendingEvents(ctx context.Context) {
	events, err := w.outboxRepo.FetchPendingEvents(ctx, 10)
	if err != nil {
		w.logger.Error("failed to fetch pending events", slog.String("error", err.Error()))
		return
	}

	for _, outboxEvent := range events {
		var taskEvent taskevent.TaskCreatedEvent
		if err := json.Unmarshal([]byte(outboxEvent.Payload), &taskEvent); err != nil {
			w.logger.Error("failed to unmarshal payload",
				slog.String("event_id", outboxEvent.ID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := w.producer.PublishTaskCreated(ctx, taskEvent.UserID.String(), taskEvent); err != nil {
			w.logger.Error("failed to publish to kafka",
				slog.String("event_id", outboxEvent.ID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := w.outboxRepo.MarkAsSent(ctx, outboxEvent.ID.String()); err != nil {
			w.logger.Error("failed to mark event as sent",
				slog.String("event_id", outboxEvent.ID.String()),
				slog.String("error", err.Error()),
			)
		}
	}
}
