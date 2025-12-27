package main

import (
	"log/slog"
	"net"
	"os"
	"strings"

	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"

	// "github.com/ThreeDotsLabs/watermill-opentelemetry/pkg/opentelemetry" REMOVED
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/username/progetto/auth/internal/handler"
	"github.com/username/progetto/auth/internal/model"
	"github.com/username/progetto/auth/internal/repository"
	"github.com/username/progetto/auth/internal/service"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	"github.com/username/progetto/shared/pkg/observability"
	"google.golang.org/grpc"
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
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
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

	// 3. Kafka Publisher
	kafkaPub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   strings.Split(kafkaBrokers, ","),
			Marshaler: kafka.DefaultMarshaler{},
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		logger.Error("failed to create kafka publisher", "error", err)
		os.Exit(1)
	}

	// Wrap with OTel (Shared)
	publisher := observability.NewTracingPublisher(kafkaPub)
	defer publisher.Close()

	// 4. Wiring
	userRepo := repository.NewPostgresRepository(db)
	tokenRepo := repository.NewRedisRepository(rdb)
	authSvc := service.NewAuthService(userRepo, tokenRepo, publisher, jwtSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	// 5. gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer(observability.GRPCServerOptions()...)
	authv1.RegisterAuthServiceServer(srv, authHandler)
	reflection.Register(srv)

	logger.Info("Auth Service gRPC server listening on :50051")
	if err := srv.Serve(lis); err != nil {
		logger.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
