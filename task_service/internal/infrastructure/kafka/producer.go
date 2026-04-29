package kafka

import (
	"context"
	"encoding/json"

	"github.com/abrshDev/task-service/internal/domain/events"
	"github.com/segmentio/kafka-go"
)

type EventProducer struct {
	writer *kafka.Writer
}

func NewEventProducer(brokers []string, topic string) *EventProducer {
	return &EventProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
			// Required for reliability: wait for all replicas to acknowledge
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *EventProducer) PublishTaskCreated(ctx context.Context, userID string, event events.TaskCreatedEvent) error {

	event.Action = "TASK_CREATED"

	messageBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(userID), // Guarantees order for this specific user
		Value: messageBytes,
	})
}

// buffered messages are flushed to the brokers.
func (p *EventProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
