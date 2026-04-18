package postgres

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("NOTIFICATION_DB_HOST"),
		os.Getenv("NOTIFICATION_DB_USER"),
		os.Getenv("NOTIFICATION_DB_PASSWORD"),
		os.Getenv("NOTIFICATION_DB_NAME"),
		os.Getenv("NOTIFICATION_DB_PORT"),
	)

	log.Printf("connection to notification DB: host %s dbname=%s", os.Getenv("NOTIFICATION_DB_HOST"), os.Getenv("NOTIFICATION_DB_NAME"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)

	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %v", err)
	}

	log.Println("Running versioned database migrations (golang-migrate)...")
	if err := RunMigrations(sqlDB); err != nil {
		return nil, fmt.Errorf("versioned migration failed: %v", err)
	}

	log.Println("Database connection established and versioned migrations applied")
	return db, nil
}
