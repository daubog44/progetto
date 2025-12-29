package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/username/progetto/shared/pkg/database/neo4j"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/social-service/internal/config"
	"github.com/username/progetto/social-service/internal/events"
	"github.com/username/progetto/social-service/internal/repository"
	"github.com/username/progetto/social-service/internal/worker"
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

	// 4. Initialize Metrics Server (Prometheus) - If needed, but we are moving to OTLP.
	// However, user might still want Prometheus port for backward compatibility or simple scraping.
	// Based on previous conversations, we are refactoring to OTLP push, so maybe not strictly needed on a separate port
	// unless the 'observability' package does it.
	// But let's follow the pattern if `InitProvider` doesn't start a server.
	// Actually, `InitProvider` usually sets up the global providers.
	// If the shared package exposes metrics via OTLP, we are good.

	// 5. Connect to Neo4j
	logger.Info("connecting to neo4j", "uri", cfg.Neo4jUri)
	driver, err := neo4j.NewNeo4j(context.Background(), cfg.Neo4jUri, cfg.Neo4jUser, cfg.Neo4jPassword)
	if err != nil {
		logger.Error("failed to connect to neo4j", "error", err)
		os.Exit(1)
	}
	defer driver.Close(context.Background())

	// 6. Setup Repository & Consumer
	neo4jRepo := repository.NewNeo4jRepository(driver)
	userConsumer := worker.NewUserConsumer(neo4jRepo)

	// 7. Setup Event Router
	router, err := events.NewEventRouter(logger, cfg.KafkaBrokers, userConsumer)
	if err != nil {
		logger.Error("failed to create event router", "error", err)
		os.Exit(1)
	}
	defer router.Close()

	// 8. Run Router
	go func() {
		if err := router.Run(context.Background()); err != nil {
			logger.Error("router failed", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("social-service started")

	// 9. Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("shutting down social-service")
	// Give some time for ongoing requests
	time.Sleep(1 * time.Second)
}
