package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/username/progetto/auth/internal/events"
	"github.com/username/progetto/auth/internal/handler"
	"github.com/username/progetto/auth/internal/model"
	"github.com/username/progetto/auth/internal/repository"
	"github.com/username/progetto/auth/internal/service"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	"github.com/username/progetto/shared/pkg/database/postgres"
	"github.com/username/progetto/shared/pkg/database/redis"
	"github.com/username/progetto/shared/pkg/grpcutil"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
	"google.golang.org/grpc/reflection"
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
	// Reset logger to default (context-aware if setup)
	logger = slog.Default()

	// Config
	dbDSN := os.Getenv("APP_DB_DSN")
	if dbDSN == "" {
		panic("APP_DB_DSN is required")
	}
	redisAddr := os.Getenv("APP_REDIS_ADDR")
	if redisAddr == "" {
		panic("APP_REDIS_ADDR is required")
	}
	kafkaBrokers := os.Getenv("APP_KAFKA_BROKERS")
	if kafkaBrokers == "" {
		panic("APP_KAFKA_BROKERS is required")
	}
	jwtSecret := os.Getenv("APP_JWT_SECRET")
	if jwtSecret == "" {
		panic("APP_JWT_SECRET is required")
	}

	// 1. Postgres
	db, err := postgres.NewPostgres(dbDSN, logger)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := postgres.AutoMigrate(db, &model.User{}); err != nil {
		slog.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	// 2. Redis
	rdb, err := redis.NewRedis(redisAddr, logger)
	if err != nil {
		slog.Error("failed to init redis", "error", err)
		os.Exit(1)
	}

	// 3. Kafka Publisher (Used by AuthService)
	publisher, err := watermillutil.NewKafkaPublisher(kafkaBrokers, logger)
	if err != nil {
		slog.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// 4. Wiring
	userRepo := repository.NewPostgresRepository(db)
	tokenRepo := repository.NewRedisRepository(rdb)
	authSvc := service.NewAuthService(userRepo, tokenRepo, publisher, jwtSecret)

	// 5. Watermill Event Router
	eventRouter, err := events.NewEventRouter(logger, kafkaBrokers, authSvc)
	if err != nil {
		slog.Error("failed to create event router", "error", err)
		os.Exit(1)
	}
	defer eventRouter.Close()

	// Start Event Router
	go func() {
		slog.Info("Starting Auth Saga Router...")
		if err := eventRouter.Run(context.Background()); err != nil {
			slog.Error("router failed", "error", err)
			// Don't os.Exit here to allow gRPC to stay up, or do?
			// Usually partial failure is bad, but maybe tolerant.
		}
	}()

	// 6. gRPC Server
	authHandler := handler.NewAuthHandler(authSvc)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpcutil.NewServer()
	authv1.RegisterAuthServiceServer(srv, authHandler)
	reflection.Register(srv)

	// Run Server
	go func() {
		slog.Info("Auth Service gRPC server listening on :50051")
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
