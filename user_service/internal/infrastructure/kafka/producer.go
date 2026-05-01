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

/*
	msg := map[string]interface{}{
			"event_type": "UserCreated",
			"user_id":    userID,
			"email":      email,
			"tenant_id":  tenantID,
		}
*/
func (p *UserProducer) PublishUserCreated(ctx context.Context, userID string, email string, tenantID uint) error {
	msg := map[string]interface{}{

		"user_id":   userID,
		"tenant_id": tenantID,
		"email":     email,
		"action":    "UserCreated",
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.Writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	})
}
