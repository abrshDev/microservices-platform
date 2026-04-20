package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type UserProducer struct {
	Writer *kafka.Writer
}

func NewUserProducer(brokers []string) *UserProducer {
	return &UserProducer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    "user-events",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *UserProducer) PublishUserCreated(ctx context.Context, userID string, email string) error {
	msg := map[string]string{
		"event_type": "UserCreated",
		"user_id":    userID,
		"email":      email,
	}
	payload, _ := json.Marshal(msg)
	return p.Writer.WriteMessages(ctx, kafka.Message{Value: payload})
}
