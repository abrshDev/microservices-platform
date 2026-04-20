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
		os.Getenv("REPORTING_DB_HOST"),
		os.Getenv("REPORTING_DB_USER"),
		os.Getenv("REPORTING_DB_PASSWORD"),
		os.Getenv("REPORTING_DB_NAME"),
		os.Getenv("REPORTING_DB_PORT"),
	)

	log.Printf("Connecting to REPORTING DB: host=%s dbname=%s", os.Getenv("REPORTING_DB_HOST"), os.Getenv("REPORTING_DB_NAME"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	// 1. Extract the underlying *sql.DB for golang-migrate
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
