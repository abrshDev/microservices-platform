package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/entities"
	"github.com/abrshDev/reporting-service/internal/domain/repositories"
	"github.com/segmentio/kafka-go"
)

type TaskEvent struct {
	UserID    uint      `json:"user_id"`
	TenantID  uint      `json:"tenant_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}
type UserEvent struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Email     string    `json:"email"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *TaskEvent) MapToEntity() entities.UserTaskSummary {
	return entities.UserTaskSummary{
		UserID:     e.UserID,
		TenantID:   e.TenantID,
		TotalTasks: 1,
		UpdatedAt:  e.Timestamp,
	}
}
func (e *UserEvent) MapToEntity() entities.UserTaskSummary {
	return entities.UserTaskSummary{
		UserID:     e.ID,
		TenantID:   e.TenantID,
		TotalTasks: 0,
		UpdatedAt:  e.Timestamp,
	}
}
func StartConsumer(brokers []string, topic string, groupID string, repo repositories.SummaryRepo) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
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

		summary := event.MapToEntity()
		if err := repo.UpsertSummary(summary); err != nil {
			log.Printf("Error updating summary in DB: %v", err)
			continue
		}

		log.Printf("Successfully processed event for User %d (Tenant %d)", event.UserID, event.TenantID)
	}
}
func StartUserConsumer(brokers []string, topic string, groupID string, repo repositories.SummaryRepo) {
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

		summary := event.MapToEntity()
		if err := repo.UpsertSummary(summary); err != nil {
			log.Printf("Error initializing user summary: %v", err)
			continue
		}
		log.Printf("Reporting Service: Created initial record for User %d", event.ID)
	}
}
