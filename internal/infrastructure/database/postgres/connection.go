package postgres

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/abrshDev/user-service/internal/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)

	var db *gorm.DB
	var err error

	// Try to connect up to 5 times
	// Try to connect up to 10 times for slow first-time starts
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Database not ready, retrying in 3 seconds... (Attempt %d/10)", i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	// Auto-Migrate
	db.AutoMigrate(&entities.User{})
	return db, nil
}
