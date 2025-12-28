package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/username/progetto/messaging-service/internal/handler"
	"github.com/username/progetto/messaging-service/internal/repository"
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

	// Config
	kafkaBrokers := os.Getenv("APP_KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	// Cassandra Config
	cassandraHost := os.Getenv("APP_CASSANDRA_HOST")
	if cassandraHost == "" {
		cassandraHost = "cassandra"
	}
	consistency := os.Getenv("APP_CASSANDRA_CONSISTENCY")
	if consistency == "" {
		consistency = "QUORUM"
	}
	// Simple env parsing for pooling options (omitted complex parsing for brevity, sticking to defaults or simple check)
	// In real app use a config library like Viper.

	// Retry loop for Cassandra connection (DNS or DB might not be ready)
	var repo repository.UserRepository
	for i := 0; i < 30; i++ {
		repo, err = repository.NewCassandraRepository(repository.CassandraConfig{
			Host:           cassandraHost,
			Consistency:    consistency,
			ConnectTimeout: 10 * time.Second,
			MaxOpenConns:   0, // Unlimited connections (user request for sharding prep)
		})
		if err == nil {
			break
		}
		slog.Info("Waiting for Cassandra...", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		slog.Error("failed to connect to cassandra after retries", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Watermill Factory Usage
	publisher, err := watermillutil.NewKafkaPublisher(kafkaBrokers, logger)
	if err != nil {
		slog.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	subscriber, err := watermillutil.NewKafkaSubscriber(kafkaBrokers, "messaging-service", logger)
	if err != nil {
		slog.Error("failed to create kafka subscriber", "error", err)
		os.Exit(1)
	}
	defer subscriber.Close()

	router, err := watermillutil.NewRouter(logger, "messaging-consumer")
	if err != nil {
		slog.Error("failed to create router", "error", err)
		os.Exit(1)
	}

	// Add Service-Specific Resiliency (Circuit Breaker) -> NOW HANDLED BY NewRouter
	// Retry is already added by NewRouter

	// Handler
	client := &handler.Client{
		Repo:      repo,
		Publisher: publisher,
	}

	router.AddConsumerHandler(
		"user_created_handler",
		"user_created",
		subscriber,
		client.HandleUserCreated,
	)

	// Run Router
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("Starting messaging-service router...")
	if err := router.Run(ctx); err != nil {
		slog.Error("router failed", "error", err)
		os.Exit(1)
	}
}
