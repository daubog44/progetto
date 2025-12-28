package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/username/progetto/auth/internal/handler"
	"github.com/username/progetto/auth/internal/model"
	"github.com/username/progetto/auth/internal/repository"
	"github.com/username/progetto/auth/internal/service"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	"github.com/username/progetto/shared/pkg/grpcutil"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	dbDSN := os.Getenv("APP_DB_DSN")
	if dbDSN == "" {
		dbDSN = "postgres://user:password@postgres:5432/auth_db?sslmode=disable"
	}
	redisAddr := os.Getenv("APP_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	kafkaBrokers := os.Getenv("APP_KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:29092"
	}
	jwtSecret := os.Getenv("APP_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretkey"
	}

	// 1. Postgres
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Connection Pooling
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get sql.DB", "error", err)
		os.Exit(1)
	}
	// Default to sensible values for production readiness
	sqlDB.SetMaxIdleConns(0)    // 0 means unlimited (keep all idle connections)
	sqlDB.SetMaxOpenConns(0)    // 0 means unlimited
	sqlDB.SetConnMaxLifetime(0) // 0 means reuse forever
	// Add OTel Gorm Plugin
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		logger.Error("failed to use otelgorm plugin", "error", err)
	}
	// Migrate
	if err := db.AutoMigrate(&model.User{}); err != nil {
		logger.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	// 2. Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Watermill Logger (Slog Adapter)
	// wLogger := observability.NewSlogWatermillAdapter(logger) // Handled inside utilities

	// 3. Kafka Publisher (Shared Factory)
	publisher, err := watermillutil.NewKafkaPublisher(kafkaBrokers, logger)
	if err != nil {
		logger.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// 4. Kafka Subscriber (For Saga)
	subscriber, err := watermillutil.NewKafkaSubscriber(kafkaBrokers, "auth-service-saga", logger)
	if err != nil {
		logger.Error("failed to create kafka subscriber", "error", err)
		os.Exit(1)
	}
	defer subscriber.Close()

	// 5. Wiring
	userRepo := repository.NewPostgresRepository(db)
	tokenRepo := repository.NewRedisRepository(rdb)
	authSvc := service.NewAuthService(userRepo, tokenRepo, publisher, jwtSecret)

	// Watermill Router (Shared Factory)
	router, err := watermillutil.NewRouter(logger, "auth-saga-consumer")
	if err != nil {
		logger.Error("failed to create router", "error", err)
		os.Exit(1)
	}

	sagaHandler := handler.NewSagaHandler(authSvc)
	router.AddConsumerHandler(
		"auth_user_creation_failed_handler",
		"user_creation_failed",
		subscriber,
		sagaHandler.HandleUserCreationFailed,
	)

	// Run Router in Background
	go func() {
		logger.Info("Starting Auth Saga Router...")
		if err := router.Run(context.Background()); err != nil {
			logger.Error("router failed", "error", err)
		}
	}()

	authHandler := handler.NewAuthHandler(authSvc)

	// 6. gRPC Server (Shared Factory)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpcutil.NewServer() // Standard options already included
	authv1.RegisterAuthServiceServer(srv, authHandler)
	reflection.Register(srv)

	logger.Info("Auth Service gRPC server listening on :50051")
	if err := srv.Serve(lis); err != nil {
		logger.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
