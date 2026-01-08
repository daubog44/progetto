package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	redis_driver "github.com/redis/go-redis/v9"
	"github.com/riandyrn/otelchi"
	"github.com/username/progetto/gateway-service/internal/api"
	"github.com/username/progetto/gateway-service/internal/config"
	"github.com/username/progetto/gateway-service/internal/events"
	"github.com/username/progetto/gateway-service/internal/sse"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	searchv1 "github.com/username/progetto/proto/gen/go/search/v1"
	"github.com/username/progetto/shared/pkg/database/redis"
	"github.com/username/progetto/shared/pkg/deduplication"
	"github.com/username/progetto/shared/pkg/grpcutil"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/watermillutil"
	"google.golang.org/grpc"
)

type App struct {
	Cfg              *config.Config
	Logger           *slog.Logger
	Router           *chi.Mux
	HumaAPI          huma.API
	WatermillManager *events.WatermillManager
	PostClient       postv1.PostServiceClient
	AuthClient       authv1.AuthServiceClient
	SearchClient     searchv1.SearchServiceClient
	SSEHandler       *sse.Handler

	// Internal connections to close
	postConn    *grpc.ClientConn
	authConn    *grpc.ClientConn
	searchConn  *grpc.ClientConn
	redisClient *redis_driver.Client
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	// 1. Redis
	rdb, err := redis.NewRedis(cfg.RedisAddr, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// 2. Deduplicator
	dedup := deduplication.NewRedisDeduplicator(rdb, "gateway:dedup")

	// 3. gRPC Clients
	postConn, err := grpcutil.NewClient(cfg.PostService, "post-service")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to post-service: %w", err)
	}
	postClient := postv1.NewPostServiceClient(postConn)

	authConn, err := grpcutil.NewClient(cfg.AuthService, "auth-service")
	if err != nil {
		postConn.Close()
		return nil, fmt.Errorf("failed to connect to auth-service: %w", err)
	}
	authClient := authv1.NewAuthServiceClient(authConn)

	searchConn, err := grpcutil.NewClient(cfg.SearchService, "search-service")
	if err != nil {
		postConn.Close()
		authConn.Close()
		return nil, fmt.Errorf("failed to connect to search-service: %w", err)
	}
	searchClient := searchv1.NewSearchServiceClient(searchConn)

	// 4. SSE Handler
	sseHandler := sse.NewHandler(rdb, cfg.JWTSecret)

	// 5. Watermill (Kafka)
	pub, err := watermillutil.NewKafkaPublisher(cfg.KafkaBrokers, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka publisher: %w", err)
	}
	sub, err := watermillutil.NewKafkaSubscriber(cfg.KafkaBrokers, "gateway-service", logger)
	if err != nil {
		pub.Close()
		return nil, fmt.Errorf("failed to create kafka subscriber: %w", err)
	}

	wmManager, err := events.NewWatermillManager(logger, pub, sub)
	if err != nil {
		pub.Close()
		sub.Close()
		return nil, fmt.Errorf("failed to create watermill manager: %w", err)
	}

	// 6. Router & Huma
	router := chi.NewMux()

	// Middlewares
	router.Use(otelchi.Middleware("gateway-service", otelchi.WithChiRoutes(router)))
	router.Use(observability.MiddlewareMetrics)
	router.Use(api.NewLoggingMiddleware(logger))
	router.Use(api.NewDeduplicationMiddleware(dedup, 10*time.Minute))
	router.Use(api.NewAdminMiddleware(cfg.JWTSecret))

	// SSE Route
	router.Get("/events", sseHandler.ServeHTTP)

	humaAPI := humachi.New(router, huma.DefaultConfig("Gateway API", "1.0.0"))

	// Register Routes
	api.RegisterPostRoutes(humaAPI, postClient, logger)
	api.RegisterAuthRoutes(humaAPI, authClient, logger)
	api.RegisterSearchRoutes(humaAPI, searchClient, logger)

	// Ping Route
	huma.Register(humaAPI, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/ping",
		Summary:     "Health check",
		Tags:        []string{"Health"},
	}, func(ctx context.Context, input *struct{}) (*struct {
		Body struct {
			Ping string `json:"ping" example:"pong"`
		}
	}, error) {
		return &struct {
			Body struct {
				Ping string `json:"ping" example:"pong"`
			}
		}{Body: struct {
			Ping string `json:"ping" example:"pong"`
		}{Ping: "pong"}}, nil
	})

	return &App{
		Cfg:              cfg,
		Logger:           logger,
		Router:           router,
		HumaAPI:          humaAPI,
		WatermillManager: wmManager,
		PostClient:       postClient,
		AuthClient:       authClient,
		SearchClient:     searchClient,
		SSEHandler:       sseHandler,
		postConn:         postConn,
		authConn:         authConn,
		searchConn:       searchConn,
		redisClient:      rdb,
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start SSE Subscriber
	go func() {
		a.Logger.Info("Starting SSE Redis Subscriber...")
		if err := a.SSEHandler.SubscribeToRedis(ctx); err != nil {
			a.Logger.Error("sse redis subscriber failed", "error", err)
		}
	}()

	// Start Event Router
	go func() {
		a.Logger.Info("Starting Gateway Event Router...")
		if err := a.WatermillManager.Router.Run(ctx); err != nil {
			a.Logger.Error("gateway router failed", "error", err)
		}
	}()

	// Start HTTP Server
	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", a.Cfg.Port),
		Handler:     a.Router,
		ReadTimeout: 5 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	go func() {
		a.Logger.Info("Starting gateway", "port", a.Cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Error("http server failed", "error", err)
		}
	}()

	<-ctx.Done()
	a.Logger.Info("Shutting down...")

	// Shutdown HTTP Server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		a.Logger.Error("server shutdown failed", "error", err)
	}

	a.Close()
	return nil
}

func (a *App) Close() {
	if a.postConn != nil {
		a.postConn.Close()
	}
	if a.authConn != nil {
		a.authConn.Close()
	}
	if a.searchConn != nil {
		a.searchConn.Close()
	}
	if a.WatermillManager != nil {
		a.WatermillManager.Close()
	}
	if a.redisClient != nil {
		a.redisClient.Close()
	}
}
