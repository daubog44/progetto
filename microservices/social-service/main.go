package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/username/progetto/shared/pkg/database/neo4j"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
	"github.com/username/progetto/social-service/internal/config"
	"github.com/username/progetto/social-service/internal/events"
	"github.com/username/progetto/social-service/internal/handler"
	"github.com/username/progetto/social-service/internal/repository"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Initialize Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 3. Initialize Observability (Tracin & Metrics)
	obsCfg := observability.LoadConfigFromEnv()
	// Override with local config if needed, but Env should match
	obsCfg.ServiceName = cfg.OtelServiceName
	obsCfg.OTLPEndpoint = cfg.OtelExporterEndpoint

	shutdownObs, err := observability.Init(context.Background(), obsCfg)
	if err != nil {
		logger.Error("failed to init observability", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownObs(context.Background()); err != nil {
			logger.Error("failed to shutdown observability", "error", err)
		}
	}()

	// 5. Connect to Neo4j
	logger.Info("connecting to neo4j", "uri", cfg.Neo4jUri)
	driver, err := neo4j.NewNeo4j(context.Background(), cfg.Neo4jUri, cfg.Neo4jUser, cfg.Neo4jPassword)
	if err != nil {
		logger.Error("failed to connect to neo4j", "error", err)
		os.Exit(1)
	}
	defer driver.Close(context.Background())

	// 6. Kafka Publisher
	publisher, err := watermillutil.NewKafkaPublisher(cfg.KafkaBrokers, logger)
	if err != nil {
		logger.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// 7. Setup Repository & Consumer
	neo4jRepo := repository.NewNeo4jRepository(driver)
	userHandler := handler.NewUserHandler(neo4jRepo, publisher)

	// 8. Setup Event Router
	router, err := events.NewEventRouter(logger, cfg.KafkaBrokers, publisher, userHandler)
	if err != nil {
		logger.Error("failed to create event router", "error", err)
		os.Exit(1)
	}
	defer router.Close()

	// 9. Standard Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 10. Run Router
	go func() {
		if err := router.Run(ctx); err != nil {
			logger.Error("router failed", "error", err)
		}
	}()

	logger.Info("social-service started")

	<-ctx.Done()
	logger.Info("shutting down social-service")
}
