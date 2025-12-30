package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5" // Imported as api package
	"github.com/username/progetto/gateway-service/internal/events"
	"github.com/username/progetto/gateway-service/internal/sse"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"github.com/username/progetto/shared/pkg/database/redis"
	"github.com/username/progetto/shared/pkg/deduplication"
	"github.com/username/progetto/shared/pkg/grpcutil"
	"github.com/username/progetto/shared/pkg/observability"
	"google.golang.org/grpc"
)

type Options struct {
	Port         int    `help:"Port to listen on" short:"p" default:"8888"`
	PostService  string `help:"Address of the post service"`
	AuthService  string `help:"Address of the auth service"`
	KafkaBrokers string `help:"Kafka brokers (comma-separated)"`
	RedisAddr    string `help:"Redis address"`
	JWTSecret    string `help:"JWT secret"`
}

type PingOutput struct {
	Body struct {
		Ping string `json:"ping" example:"pong" doc:"Response from gateway-service"`
	}
}

func LoadEnv(options *Options) {

	if envPost := os.Getenv("POST_SERVICE"); envPost != "" {
		options.PostService = envPost
	}
	if envAuth := os.Getenv("AUTH_SERVICE"); envAuth != "" {
		options.AuthService = envAuth
	}
	if envKafka := os.Getenv("KAFKA_BROKERS"); envKafka != "" {
		options.KafkaBrokers = envKafka
	}
	if envRedis := os.Getenv("APP_REDIS_ADDR"); envRedis != "" {
		options.RedisAddr = envRedis
	}
	if envJWT := os.Getenv("APP_JWT_SECRET"); envJWT != "" {
		options.JWTSecret = envJWT
	}

	// Strict Check
	if options.PostService == "" {
		panic("POST_SERVICE is required")
	}
	if options.AuthService == "" {
		panic("AUTH_SERVICE is required")
	}
	if options.KafkaBrokers == "" {
		panic("KAFKA_BROKERS is required")
	}
	if options.RedisAddr == "" {
		panic("APP_REDIS_ADDR is required")
	}
	if options.JWTSecret == "" {
		panic("APP_JWT_SECRET is required")
	}

}

func initGrpcClientPost(logger *slog.Logger, options *Options) (*grpc.ClientConn, postv1.PostServiceClient) {
	// gRPC Client setup: Post Service
	postConn, err := grpcutil.NewClient(options.PostService, "post-service")
	if err != nil {
		logger.Error("failed to connect to post-service", "error", err)
		os.Exit(1)
	}
	postClient := postv1.NewPostServiceClient(postConn)

	return postConn, postClient
}

func initGrpcClientAuth(logger *slog.Logger, options *Options) (*grpc.ClientConn, authv1.AuthServiceClient) {
	// gRPC Client setup: Auth Service

	// example with retry middlewareS
	// retryOpts := grpcutil.DefaultRetryOptions()
	// grpc.WithUnaryInterceptor(grpcutil.SmartRetryUnaryClientInterceptor(retryOpts))
	authConn, err := grpcutil.NewClient(options.AuthService, "auth-service")
	if err != nil {
		logger.Error("failed to connect to auth-service", "error", err)
		os.Exit(1)
	}
	authClient := authv1.NewAuthServiceClient(authConn)

	return authConn, authClient
}

func CLI(logger *slog.Logger, httpAPI huma.API) humacli.CLI {
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Manual environment variable overrides (Docker compatibility)
		LoadEnv(options)

		// Redis Setup
		rdb, err := redis.NewRedis(options.RedisAddr, logger)
		if err != nil {
			logger.Error("failed to connect to redis", "error", err)
			os.Exit(1)
		}

		// Deduplicator
		dedup := deduplication.NewRedisDeduplicator(rdb, "gateway:dedup")

		// Create a new router & API
		router := chi.NewMux()
		// Middlewares
		router.Use(observability.MiddlewareMetrics)
		router.Use(NewLoggingMiddleware(logger)) // Added Logging Middleware

		// CORS
		router.Use(NewDeduplicationMiddleware(dedup, 10*time.Minute))
		router.Use(NewAdminMiddleware(options.JWTSecret))

		// SSE Handler
		sseHandler := sse.NewHandler(rdb, options.JWTSecret)
		router.Get("/events", sseHandler.ServeHTTP)

		// Start SSE Redis Subscriber
		go func() {
			logger.Info("Starting SSE Redis Subscriber...")
			if err := sseHandler.SubscribeToRedis(context.Background()); err != nil {
				logger.Error("sse redis subscriber failed", "error", err)
			}
		}()

		httpAPI = humachi.New(router, huma.DefaultConfig("Gateway API", "1.0.0"))

		// gRPC Client setup: Post Service
		postConn, postClient := initGrpcClientPost(logger, options)

		// gRPC Client setup: Auth Service
		authConn, authClient := initGrpcClientAuth(logger, options)

		// Watermill Event Router
		eventRouter, err := events.NewEventRouter(logger, options.KafkaBrokers, sseHandler)
		if err != nil {
			logger.Error("failed to create event router", "error", err)
			os.Exit(1)
		}

		// Start Event Router in Background
		go func() {
			logger.Info("Starting Gateway Event Router...")
			if err := eventRouter.Run(context.Background()); err != nil {
				logger.Error("gateway router failed", "error", err)
			}
		}()

		// Register Routes
		RegisterPostRoutes(httpAPI, postClient, logger)
		RegisterAuthRoutes(httpAPI, authClient, logger)

		// Register GET /ping handler.
		huma.Register(httpAPI, huma.Operation{
			OperationID: "health-check",
			Method:      http.MethodGet,
			Path:        "/ping",
			Summary:     "Health check",
			Tags:        []string{"Health"},
		}, func(ctx context.Context, input *struct{}) (*PingOutput, error) {
			resp := &PingOutput{}
			resp.Body.Ping = "pong"
			return resp, nil
		})

		// Admin Route Example
		huma.Register(httpAPI, huma.Operation{
			OperationID: "admin-check",
			Method:      http.MethodGet,
			Path:        "/admin/check",
			Summary:     "Admin check",
			Tags:        []string{"Admin"},
		}, func(ctx context.Context, input *struct{}) (*struct {
			Body struct {
				Message string `json:"message"`
			}
		}, error) {
			return &struct {
				Body struct {
					Message string `json:"message"`
				}
			}{
				Body: struct {
					Message string `json:"message"`
				}{Message: "You are admin!"},
			}, nil
		})

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			server := &http.Server{
				Addr:        fmt.Sprintf(":%d", options.Port),
				Handler:     router,
				ReadTimeout: 5 * time.Second,
				// WriteTimeout: 10 * time.Second, // Disabled for SSE support
				IdleTimeout: 120 * time.Second,
			}
			logger.Info("Starting gateway", "port", options.Port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("http server failed", "error", err)
				os.Exit(1)
			}
		})
		hooks.OnStop(func() {
			postConn.Close()
			authConn.Close()
			eventRouter.Close()
		})
	})

	return cli
}
