package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/repositories"
	"github.com/segmentio/kafka-go"
)

type TaskEvent struct {
	ID        string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	TenantID  uint64    `json:"tenant_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}
type UserEvent struct {
	UserID    string    `json:"user_id"`
	TenantID  uint64    `json:"tenant_id"`
	Email     string    `json:"email"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

func StartTaskConsumer(brokers []string, topic string, groupID string, repo repositories.SummaryRepo, ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer reader.Close()

	log.Printf("Kafka Consumer started: Topic=%s, Group=%s", topic, groupID)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var event TaskEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		// IDEMPOTENCY CHECK
		isNew, err := repo.InsertIfNotExist(ctx, event.ID)
		if err != nil {
			log.Printf("failed to check idempotency: %v", err)
			continue
		}
		if !isNew {
			log.Printf("duplicate event detected, skipping: %s", event.ID)
			continue // Skip this message, keep consumer running
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
				log.Printf("Failed to process audit update: %v", err)
				continue
			}
		}

		// Mark as completed AFTER successful processing
		repo.UpdateStatus(ctx, event.ID, "COMPLETED")

		log.Printf("Processed %s for User %s", event.Action, event.UserID)
	}
}

func StartUserConsumer(brokers []string, topic string, groupID string, repo repositories.SummaryRepo, ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	defer reader.Close()

	log.Printf("Kafka User Consumer started: Topic=%s", topic)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading user message: %v", err)
			continue
		}

		var event UserEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Error unmarshaling user event: %v", err)
			continue
		}
		isNew, err := repo.InsertIfNotExist(ctx, event.UserID)
		if err != nil {
			log.Printf("failed to check idempotency: %v", err)
			continue
		}
		if !isNew {
			log.Printf("duplicate event detected, skipping: %s", event.UserID)
			continue
		}
		err = repo.UpdateWithAudit(event.UserID, event.TenantID, 0, "USER_REGISTERED")
		if err != nil {
			log.Printf("Error initializing user summary with audit: %v", err)
			continue
		}
		repo.UpdateStatus(ctx, event.UserID, "COMPLETED")
		log.Printf("Reporting Service: Created initial record for User %s", event.UserID)
	}
}
