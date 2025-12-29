package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/username/progetto/messaging-service/internal/events"
	"github.com/username/progetto/messaging-service/internal/handler"
	"github.com/username/progetto/messaging-service/internal/repository"
	"github.com/username/progetto/shared/pkg/database/cassandra"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Init Observability
	obsCfg := observability.LoadConfigFromEnv()
	shutdown, err := observability.Init(context.Background(), obsCfg)
	if err != nil {
		slog.Error("failed to init observability", "error", err)
	}
	defer func() {
		if shutdown != nil {
			shutdown(context.Background())
		}
	}()

	// Config - Strict Mode (Panic if missing)
	kafkaBrokers := os.Getenv("APP_KAFKA_BROKERS")
	if kafkaBrokers == "" {
		panic("APP_KAFKA_BROKERS environment variable is not set")
	}
	cassandraHost := os.Getenv("APP_CASSANDRA_HOST")
	if cassandraHost == "" {
		panic("APP_CASSANDRA_HOST environment variable is not set")
	}
	// Consistency can have a default if needed, or strict. Let's keep default for now as it's less critical?
	// User said "sostituisci tutte le parti in cui ha inserito una logica simile".
	// Let's enforce strictness where reasonable or critical.
	consistency := os.Getenv("APP_CASSANDRA_CONSISTENCY")
	if consistency == "" {
		consistency = "QUORUM" // Default is acceptable here as it's a tuning param
	}

	cassandraKeyspace := os.Getenv("APP_CASSANDRA_KEYSPACE")
	if cassandraKeyspace == "" {
		panic("APP_CASSANDRA_KEYSPACE environment variable is not set")
	}

	// Retry loop for Cassandra connection
	// Retry loop for Cassandra connection
	session, err := cassandra.NewCassandra(cassandra.Config{
		Host:           cassandraHost,
		Consistency:    consistency,
		ConnectTimeout: 10 * time.Second,
		Keyspace:       cassandraKeyspace,
	}, logger)
	if err != nil {
		slog.Error("failed to connect to cassandra", "error", err)
		os.Exit(1)
	}
	defer session.Close()

	repo := repository.NewCassandraRepository(session)

	// Watermill Publisher (Shared Factory)
	publisher, err := watermillutil.NewKafkaPublisher(kafkaBrokers, logger)
	if err != nil {
		slog.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// Handler
	client := &handler.Client{
		Repo:      repo,
		Publisher: publisher,
		Logger:    logger.With("component", "messaging_client"),
	}

	// Watermill Event Router
	eventRouter, err := events.NewEventRouter(logger, kafkaBrokers, client)
	if err != nil {
		slog.Error("failed to create event router", "error", err)
		os.Exit(1)
	}
	defer eventRouter.Close()

	// Run Router
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("Starting messaging-service router...")
	if err := eventRouter.Run(ctx); err != nil {
		slog.Error("router failed", "error", err)
		os.Exit(1)
	}
}
