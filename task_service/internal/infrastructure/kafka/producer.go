package kafka

import (
	"context"
	"encoding/json"

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

func (p *EventProducer) PublishTaskCreated(ctx context.Context, event interface{}) error {
	messageBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
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
