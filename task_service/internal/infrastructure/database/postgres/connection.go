package postgres

import (
	"fmt"
	"log"
	"os"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection() (*gorm.DB, error) {
	// FIX: Changed to TASK_ prefixed variables to match your .env/compose
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("TASK_DB_HOST"),
		os.Getenv("TASK_DB_USER"),
		os.Getenv("TASK_DB_PASSWORD"),
		os.Getenv("TASK_DB_NAME"),
		os.Getenv("TASK_DB_PORT"),
	)

	log.Printf("Connecting to Task DB: host=%s dbname=%s", os.Getenv("TASK_DB_HOST"), os.Getenv("TASK_DB_NAME"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	log.Println("Running database migrations...")
	err = db.AutoMigrate(&entities.Task{})
	if err != nil {
		return nil, fmt.Errorf("migration failed: %v", err)
	}

	log.Println("Database connection established and migrated")
	return db, nil
}
