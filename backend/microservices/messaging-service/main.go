package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/username/progetto/messaging-service/internal/config"
	"github.com/username/progetto/messaging-service/internal/events"
	"github.com/username/progetto/messaging-service/internal/handler"
	"github.com/username/progetto/messaging-service/internal/repository"
	"github.com/username/progetto/shared/pkg/database/cassandra"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

func main() {
	// 0. Load Config
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Init Observability
	obsCfg := observability.LoadConfigFromEnv()
	obsCfg.ServiceName = cfg.OtelServiceName
	obsCfg.OTLPEndpoint = cfg.OtelExporterEndpoint

	shutdown, err := observability.Init(context.Background(), obsCfg)
	if err != nil {
		slog.Error("failed to init observability", "error", err)
	}
	defer func() {
		if shutdown != nil {
			shutdown(context.Background())
		}
	}()

	// 1. Connect to Cassandra
	session, err := cassandra.NewCassandra(cassandra.Config{
		Host:           cfg.CassandraHost,
		Consistency:    cfg.CassandraConsistency,
		ConnectTimeout: 10 * time.Second,
		Keyspace:       cfg.CassandraKeyspace,
	}, logger)
	if err != nil {
		slog.Error("failed to connect to cassandra", "error", err)
		os.Exit(1)
	}
	defer session.Close()

	repo := repository.NewCassandraRepository(session)

	// Watermill Publisher (Shared Factory)
	publisher, err := watermillutil.NewKafkaPublisher(cfg.KafkaBrokers, logger)
	if err != nil {
		slog.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// Handler
	handler := &handler.Handler{
		Repo:      repo,
		Publisher: publisher,
		Logger:    logger.With("component", "messaging_handler"),
	}

	// Watermill Event Router
	eventRouter, err := events.NewEventRouter(logger, cfg.KafkaBrokers, handler)
	if err != nil {
		slog.Error("failed to create event router", "error", err)
		os.Exit(1)
	}
	defer eventRouter.Close()

	// Standard Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("Starting messaging-service router...")
	if err := eventRouter.Run(ctx); err != nil {
		slog.Error("router failed", "error", err)
	}
}
