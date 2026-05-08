package main

import (
	"context"
	"log"
	"os"

	"github.com/abrshDev/reporting-service/internal/app/report/queries"
	"github.com/abrshDev/reporting-service/internal/delivery/http"
	"github.com/abrshDev/reporting-service/internal/delivery/http/handlers"
	"github.com/abrshDev/reporting-service/internal/infrastructure/config"
	"github.com/abrshDev/reporting-service/internal/infrastructure/database/postgres"
	"github.com/abrshDev/reporting-service/internal/infrastructure/kafka"
	"github.com/abrshDev/reporting-service/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	appLogger := logger.NewLogger("reporting_service")
	config.LoadEnv()

	db, err := postgres.NewConnection()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	summaryRepo := postgres.NewSummaryRepository(db)

	// Setup Queries
	getSummaryQuery := queries.NewGetSummaryQuery(summaryRepo)
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:29092"
	}
	consumerCtx, cancelConsumers := context.WithCancel(context.Background())
	defer cancelConsumers()
	// Kafka Consumer
	go func() {

		kafka.StartTaskConsumer([]string{brokers}, "task-events", "reporting-group", summaryRepo, consumerCtx, appLogger)
	}()

	go func() {
		kafka.StartUserConsumer([]string{brokers}, "user-events", "reporting-user-group", summaryRepo, consumerCtx, appLogger)
	}()

	app := fiber.New()

	// Setup Handlers and Router
	reportHandler := handlers.NewReportHandler(getSummaryQuery)
	http.SetupRoutes(app, reportHandler)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	log.Fatal(app.Listen(":8083"))
}
