package main

import (
	"fmt"
	"log"

	"github.com/abrshDev/reporting-service/internal/infrastructure/config"
	"github.com/abrshDev/reporting-service/internal/infrastructure/database/postgres"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load variables from .env
	config.LoadEnv()

	// Connect to DB and run migrations
	db, err := postgres.NewConnection()
	fmt.Println("db:", db)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	// Simple Fiber app to check status
	app := fiber.New()

	// Basic route to verify service is alive
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Reporting Service is Online")
	})

	// Start server on port 8083
	log.Println("Reporting REST server starting on :8083")
	log.Fatal(app.Listen(":8083"))
}
