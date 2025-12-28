package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/gateway-service/internal/sse"
	"github.com/username/progetto/shared/pkg/grpcutil"
	"github.com/username/progetto/shared/pkg/watermillutil"

	// "github.com/ThreeDotsLabs/watermill-opentelemetry/pkg/opentelemetry" REMOVED
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"github.com/username/progetto/shared/pkg/observability"

	"context"
	"encoding/json"

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
	Port         int    `help:"Port to listen on" short:"p" default:"8888"`
	PostService  string `help:"Address of the post service" default:"post-service:50051"`
	AuthService  string `help:"Address of the auth service" default:"auth-service:50051"`
	KafkaBrokers string `help:"Kafka brokers (comma-separated)" default:"kafka:29092"`
}

func main() {
	var api huma.API

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

		// Create a new router & API
		router := chi.NewMux()
		// Middleware
		router.Use(observability.Middleware)
		router.Use(AdminMiddleware)

		// SSE Handler
		sseHandler := sse.NewHandler()
		router.Get("/events", sseHandler.ServeHTTP)

		api = humachi.New(router, huma.DefaultConfig("Gateway API", "1.0.0"))

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

		// Watermill Publisher (Shared Factory)
		// wLogger := observability.NewSlogWatermillAdapter(logger) // Handled

		publisher, err := watermillutil.NewKafkaPublisher(options.KafkaBrokers, logger)
		if err != nil {
			logger.Error("failed to create watermill publisher", "error", err)
			os.Exit(1)
		}
		defer publisher.Close()

		// Kafka Subscriber (For SSE events from backend)
		// Using "gateway-service-sse" group.
		subscriber, err := watermillutil.NewKafkaSubscriber(options.KafkaBrokers, "gateway-service-sse", logger)
		if err != nil {
			logger.Error("failed to create watermill subscriber", "error", err)
			os.Exit(1)
		}
		defer subscriber.Close()

		// Router for Gateway Inbound Events
		msgRouter, err := watermillutil.NewRouter(logger, "gateway-events")
		if err != nil {
			logger.Error("failed to create router", "error", err)
			os.Exit(1)
		}

		// Handler: Broadcast to SSE
		broadcastHandler := func(msg *message.Message) error {
			// user_created payload: {user_id, ...}
			// user_creation_failed payload: {user_id, reason}
			// Common: has user_id
			var payload struct {
				UserID string `json:"user_id"`
			}
			// Best effort unmarshal to get ID
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				return nil // skip
			}

			// Topic name as event type
			topic := message.SubscribeTopicFromCtx(msg.Context())

			sseHandler.Broadcast(msg.Context(), payload.UserID, topic, string(msg.Payload))
			return nil
		}

		msgRouter.AddConsumerHandler("gateway_user_created", "user_created", subscriber, broadcastHandler)
		msgRouter.AddConsumerHandler("gateway_user_creation_failed", "user_creation_failed", subscriber, broadcastHandler)

		go func() {
			slog.Info("Starting Gateway Event Router...")
			if err := msgRouter.Run(context.Background()); err != nil {
				slog.Error("gateway router failed", "error", err)
			}
		}()

		// Register Routes
		RegisterPostRoutes(api, postClient, logger)
		RegisterAuthRoutes(api, authClient, logger)

		// Register GET /ping handler.
		huma.Register(api, huma.Operation{
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
		huma.Register(api, huma.Operation{
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
			logger.Info("Starting gateway", "port", options.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
		})
		hooks.OnStop(func() {
			postConn.Close()
			authConn.Close()
			publisher.Close()
			msgRouter.Close()
		})
	})

	cli.Root().AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI spec",
		Run: func(cmd *cobra.Command, args []string) {
			b, _ := api.OpenAPI().DowngradeYAML()
			logger.Info("OpenAPI Spec", "content", string(b))
		},
	})

	cli.Run()
}
