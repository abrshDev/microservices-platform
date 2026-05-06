package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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
		"postgres:15-alpine",
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

	if err := db.AutoMigrate(
		&entities.UserTaskSummary{},
		&entities.AuditLog{},
	); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS processed_events (
			event_id UUID PRIMARY KEY,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
			processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		t.Fatalf("failed to create processed_events table: %v", err)
	}

	repo := postgreslocal.NewSummaryRepository(db)

	t.Log("database ready, repo created")

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/cp-kafka:7.4.0",
		kafka.WithClusterID("test-cluster-123"),
	)
	if err != nil {
		t.Fatalf("failed to start kafka: %v", err)
	}
	defer kafkaContainer.Terminate(ctx)

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("failed to get kafka brokers: %v", err)
	}
	t.Logf("kafka brokers: %v", brokers)
	topic := "taskevents"

	if err := createTopic(ctx, brokers[0], topic); err != nil {
		t.Fatalf("failed to create topic %q: %v", topic, err)
	}

	consumerCtx, cancel := context.WithCancel(ctx)
	consumerDone := make(chan struct{})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	go func() {
		defer close(consumerDone)
		StartTaskConsumer(brokers, topic, "task-consumer-test", repo, consumerCtx, logger)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-consumerDone:
		case <-time.After(10 * time.Second):
			t.Errorf("task consumer did not stop before timeout")
		}
	})

	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers[0]),
		Topic:        topic,
		RequiredAcks: kafkago.RequireAll,
	}
	defer writer.Close()

	testEvent := TaskEvent{
		ID:            "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		UserID:        "11111111-2222-3333-4444-555555555555",
		TenantID:      1,
		Action:        "TASK_CREATED",
		CorrelationID: "test-correlation-123",
		Timestamp:     time.Now().UTC(),
	}

	eventBytes, err := json.Marshal(testEvent)
	if err != nil {
		t.Fatalf("failed to marshal test event: %v", err)
	}

	err = writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(testEvent.ID),
		Value: eventBytes,
	})
	if err != nil {
		t.Fatalf("failed to write test event: %v", err)
	}
	t.Log("test event sent to kafka")

	if err := waitForCondition(20*time.Second, 200*time.Millisecond, func() (bool, error) {
		var summary entities.UserTaskSummary
		summaryTx := db.WithContext(ctx).
			Raw(
				"SELECT user_id, tenant_id, total_tasks, updated_at FROM user_task_summaries WHERE user_id = ? AND tenant_id = ? LIMIT 1",
				testEvent.UserID,
				testEvent.TenantID,
			).
			Scan(&summary)
		if summaryTx.Error != nil {
			return false, summaryTx.Error
		}

		if summaryTx.RowsAffected == 0 || summary.TotalTasks != 1 {
			return false, nil
		}

		var processedEvent struct {
			Status string
		}
		tx := db.WithContext(ctx).
			Raw("SELECT status FROM processed_events WHERE event_id = ?", testEvent.ID).
			Scan(&processedEvent)
		if tx.Error != nil {
			return false, tx.Error
		}

		return tx.RowsAffected == 1 && processedEvent.Status == "COMPLETED", nil
	}); err != nil {
		t.Fatalf("consumer did not process the event: %v", err)
	}

	summary, err := repo.GetSummary(testEvent.UserID, testEvent.TenantID)
	if err != nil {
		t.Fatalf("failed to fetch summary after processing: %v", err)
	}
	if summary.TotalTasks != 1 {
		t.Fatalf("expected total_tasks to be 1, got %d", summary.TotalTasks)
	}

	var auditCount int64
	if err := db.Model(&entities.AuditLog{}).
		Where("user_id = ? AND tenant_id = ? AND action_type = ?", testEvent.UserID, testEvent.TenantID, testEvent.Action).
		Count(&auditCount).Error; err != nil {
		t.Fatalf("failed to count audit records: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected 1 audit log entry, got %d", auditCount)
	}
}

func createTopic(ctx context.Context, broker string, topic string) error {
	var lastErr error
	deadline := time.Now().Add(15 * time.Second)

	for time.Now().Before(deadline) {
		conn, err := kafkago.DialContext(ctx, "tcp", broker)
		if err == nil {
			err = conn.CreateTopics(kafkago.TopicConfig{
				Topic:             topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			})
			if closeErr := conn.Close(); err == nil && closeErr != nil {
				err = closeErr
			}
		}

		if err == nil {
			return nil
		}

		lastErr = err

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}

	return lastErr
}

func waitForCondition(timeout time.Duration, interval time.Duration, fn func() (bool, error)) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		ok, err := fn()
		if err != nil {
			lastErr = err
		}
		if ok {
			return nil
		}

		time.Sleep(interval)
	}

	if lastErr != nil {
		return lastErr
	}

	return errors.New("timed out waiting for condition")
}
