package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/username/progetto/auth/internal/config"
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
	// Reset logger to default (context-aware if setup)
	logger = slog.Default()

	// 1. Postgres
	db, err := postgres.NewPostgres(cfg.DbDSN, logger)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := postgres.AutoMigrate(db, &model.User{}); err != nil {
		slog.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	// 2. Redis
	rdb, err := redis.NewRedis(cfg.RedisAddr, logger)
	if err != nil {
		slog.Error("failed to init redis", "error", err)
		os.Exit(1)
	}

	// 3. Kafka Publisher (Used by AuthService)
	publisher, err := watermillutil.NewKafkaPublisher(cfg.KafkaBrokers, logger)
	if err != nil {
		slog.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// 4. Wiring
	userRepo := repository.NewPostgresRepository(db)
	tokenRepo := repository.NewRedisRepository(rdb)
	authSvc := service.NewAuthService(userRepo, tokenRepo, publisher, cfg.JwtSecret)

	// 5. Watermill Event Router
	eventRouter, err := events.NewEventRouter(logger, cfg.KafkaBrokers, authSvc)
	if err != nil {
		slog.Error("failed to create event router", "error", err)
		os.Exit(1)
	}
	defer eventRouter.Close()

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

	// Standard Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start Event Router
	go func() {
		slog.Info("Starting Auth Saga Router...")
		if err := eventRouter.Run(ctx); err != nil {
			slog.Error("router failed", "error", err)
		}
	}()

	// Run Server
	go func() {
		slog.Info("Auth Service gRPC server listening on :50051")
		if err := srv.Serve(lis); err != nil {
			slog.Error("failed to serve", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down auth-service...")
	srv.GracefulStop()
}
