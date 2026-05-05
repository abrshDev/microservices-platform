package kafka

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/abrshDev/reporting-service/internal/domain/entities"
	_ "github.com/lib/pq"

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
		time.Sleep(2 * time.Second)
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
}
