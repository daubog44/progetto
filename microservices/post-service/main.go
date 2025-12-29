package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/username/progetto/post-service/internal/events"
	"github.com/username/progetto/post-service/internal/handler"
	"github.com/username/progetto/post-service/internal/repository"
	"github.com/username/progetto/post-service/internal/worker"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"github.com/username/progetto/shared/pkg/database/mongo"
	"github.com/username/progetto/shared/pkg/grpcutil"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
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
	logger = slog.Default()

	// Config
	mongoURI := os.Getenv("APP_MONGO_URI")
	if mongoURI == "" {
		panic("APP_MONGO_URI is required")
	}
	kafkaBrokers := os.Getenv("APP_KAFKA_BROKERS")
	if kafkaBrokers == "" {
		panic("APP_KAFKA_BROKERS is required")
	}

	// 1. MongoDB
	client, db, err := mongo.NewMongo(context.Background(), mongoURI, "progetto")
	if err != nil {
		slog.Error("failed to connect to mongodb", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error("failed to disconnect mongodb", "error", err)
		}
	}()

	// 2. Repositories
	postRepo := repository.NewMongoPostRepository(db)
	userRepo := repository.NewMongoUserRepository(db)

	// 3. Kafka Publisher (Shared)
	publisher, err := watermillutil.NewKafkaPublisher(kafkaBrokers, logger)
	if err != nil {
		slog.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// 4. Wiring
	userConsumer := worker.NewUserConsumer(userRepo)
	postHandler := handler.NewPostHandler(postRepo, publisher)

	// 5. Watermill Event Router (User Sync)
	eventRouter, err := events.NewEventRouter(logger, kafkaBrokers, userConsumer)
	if err != nil {
		slog.Error("failed to create event router", "error", err)
		os.Exit(1)
	}
	defer eventRouter.Close()

	// Start Event Router
	go func() {
		slog.Info("Starting Post Service Event Router...")
		if err := eventRouter.Run(context.Background()); err != nil {
			slog.Error("router failed", "error", err)
		}
	}()

	// 6. gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpcutil.NewServer()
	postv1.RegisterPostServiceServer(srv, postHandler)
	reflection.Register(srv)

	// Run Server
	go func() {
		slog.Info("Post Service gRPC server listening on :50051")
		if err := srv.Serve(lis); err != nil {
			slog.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	slog.Info("Shutting down...")
	srv.GracefulStop()
}
