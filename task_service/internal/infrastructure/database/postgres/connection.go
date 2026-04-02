package postgres

import (
	"fmt"
	"log"
	"os"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewConnection initializes the Postgres DB and runs migrations
func NewConnection() (*gorm.DB, error) {
	// Pulling from .env (which will be provided by Docker Compose)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	// Senior Move: Auto-Migrate the entities defined in the Domain
	log.Println("Running database migrations...")
	err = db.AutoMigrate(&entities.Task{})
	if err != nil {
		return nil, fmt.Errorf("migration failed: %v", err)
	}

	log.Println("Database connection established and migrated")
	return db, nil
}
