package main

import (
	"log"

	"github.com/abrshDev/user-service/internal/app/user/commands"
	"github.com/abrshDev/user-service/internal/app/user/queries"
	"github.com/abrshDev/user-service/internal/delivery/http" // Import our new router
	"github.com/abrshDev/user-service/internal/delivery/http/handlers"
	"github.com/abrshDev/user-service/internal/infrastructure/database/postgres"
	"github.com/gofiber/fiber/v2"
)

func main() {
	db, _ := postgres.NewConnection()
	userRepo := postgres.NewUserRepository(db)

	createUserCmd := commands.NewCreateUserHandler(userRepo)
	getUserQuery := queries.NewGetUserHandler(userRepo)
	DeleteUserCmd := commands.NewDeleteUserHandler(userRepo)
	LoginQuery := queries.NewLoginHandler(userRepo)

	userHttpHandler := handlers.NewUserHandler(createUserCmd, getUserQuery, DeleteUserCmd, LoginQuery)

	app := fiber.New()

	http.SetupRoutes(app, userHttpHandler)

	log.Fatal(app.Listen(":8080"))
}
