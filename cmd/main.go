package main

import (
	"log"

	"github.com/abrshDev/user-service/internal/infrastructure/database/postgres"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Initialize Database with GORM
	// This also runs AutoMigrate (creates the 'users' table automatically)
	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected and Migrated!", db)

	// 2. Initialize Fiber
	app := fiber.New(fiber.Config{
		AppName: "User Service v1.0",
	})

	// 3. Health Check Route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status":  "success",
			"message": "User Service is running",
		})
	})

	// 4. Start the Server
	log.Fatal(app.Listen(":8080"))
}
