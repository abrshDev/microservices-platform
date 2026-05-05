package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/entities"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"

	postgreslocal "github.com/abrshDev/reporting-service/internal/infrastructure/database/postgres"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTaskConsumer_ProcessesEvent(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("reporting_db"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Wait for Postgres to be ready using database/sql
	var sqlDB *sql.DB
	for i := 0; i < 10; i++ {
		sqlDB, err = sql.Open("postgres", connStr)
		if err == nil {
			err = sqlDB.Ping()
			if err == nil {
				break
			}
		}
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		t.Fatalf("postgres not ready: %v", err)
	}
	sqlDB.Close()

	// Now connect with GORM
	db, err := gorm.Open(pg.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	db.AutoMigrate(
		&entities.UserTaskSummary{},
		&entities.AuditLog{},
	)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS processed_events (
			event_id UUID PRIMARY KEY,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
			processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)

	repo := postgreslocal.NewSummaryRepository(db)

	t.Log("database ready, repo created")
	_ = repo

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/cp-kafka:7.4.0",
		kafka.WithClusterID("test-cluster-123"),
	)
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("failed to get kafka brokers: %v", err)
	}
	t.Logf("kafka brokers: %v", brokers)
	topic := "taskevents"

	writer := &kafkago.Writer{
		Addr:  kafkago.TCP(brokers[0]),
		Topic: topic,
	}
	defer writer.Close()
	testEvent := map[string]interface{}{
		"task_id":        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		"user_id":        "11111111-2222-3333-4444-555555555555",
		"tenant_id":      uint64(1),
		"action":         "TASK_CREATED",
		"correlation_id": "test-correlation-123",
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	eventBytes, _ := json.Marshal(testEvent)
	err = writer.WriteMessages(ctx, kafkago.Message{
		Value: eventBytes,
	})
	if err != nil {
		t.Fatalf("failed to write test event: %v", err)
	}
	t.Log("test event sent to kafka")
}
