package main

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
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"github.com/username/progetto/gateway-service/internal/api"
	"github.com/username/progetto/gateway-service/internal/events"
	"github.com/username/progetto/gateway-service/internal/sse"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"github.com/username/progetto/shared/pkg/grpcutil"
	"github.com/username/progetto/shared/pkg/observability"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

// Ping pong request
type PingOutput struct {
	Body struct {
		Ping string `json:"ping" example:"pong" doc:"Response from gateway-service"`
	}
}

// Options for the CLI.
type Options struct {
	Port         int    `help:"Port to listen on" short:"p" default:"8888"` // Port default is usually fine for local dev? User said "panic se non c'è". Let's remove default for services addresses.
	PostService  string `help:"Address of the post service"`
	AuthService  string `help:"Address of the auth service"`
	KafkaBrokers string `help:"Kafka brokers (comma-separated)"`
}

func main() {
	var httpAPI huma.API

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

	// Use context-aware logger initialized by observability
	logger = slog.Default()

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Manual environment variable overrides (Docker compatibility)
		if envPost := os.Getenv("POST_SERVICE"); envPost != "" {
			options.PostService = envPost
		}
		if envAuth := os.Getenv("AUTH_SERVICE"); envAuth != "" {
			options.AuthService = envAuth
		}
		if envKafka := os.Getenv("KAFKA_BROKERS"); envKafka != "" {
			options.KafkaBrokers = envKafka
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

		// Create a new router & API
		router := chi.NewMux()
		// Middleware
		router.Use(observability.Middleware)
		router.Use(api.AdminMiddleware) // Moved to api package

		// SSE Handler
		sseHandler := sse.NewHandler()
		router.Get("/events", sseHandler.ServeHTTP)

		httpAPI = humachi.New(router, huma.DefaultConfig("Gateway API", "1.0.0"))

		// gRPC Client setup: Post Service
		postConn, err := grpcutil.NewClient(options.PostService, "post-service")
		if err != nil {
			logger.Error("failed to connect to post-service", "error", err)
			os.Exit(1)
		}
		postClient := postv1.NewPostServiceClient(postConn)

		// gRPC Client setup: Auth Service
		authConn, err := grpcutil.NewClient(options.AuthService, "auth-service")
		if err != nil {
			logger.Error("failed to connect to auth-service", "error", err)
			os.Exit(1)
		}
		authClient := authv1.NewAuthServiceClient(authConn)

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
		api.RegisterPostRoutes(httpAPI, postClient, logger)
		api.RegisterAuthRoutes(httpAPI, authClient, logger)

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
				Addr:         fmt.Sprintf(":%d", options.Port),
				Handler:      router,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
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

	cli.Root().AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI spec",
		Run: func(cmd *cobra.Command, args []string) {
			b, _ := httpAPI.OpenAPI().DowngradeYAML()
			logger.Info("OpenAPI Spec", "content", string(b))
		},
	})

	cli.Run()
}
