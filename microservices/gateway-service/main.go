package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"

	// "github.com/ThreeDotsLabs/watermill-opentelemetry/pkg/opentelemetry" REMOVED
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	authv1 "github.com/username/progetto/proto/gen/go/auth/v1"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
	"github.com/username/progetto/shared/pkg/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"context"

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
		// Prometheus endpoint
		router.Handle("/metrics", observability.PrometheusHandler())

		api = humachi.New(router, huma.DefaultConfig("Gateway API", "1.0.0"))

		// gRPC Client setup: Post Service
		postConn, err := grpc.NewClient(options.PostService, append(observability.GRPCClientOptions(), grpc.WithTransportCredentials(insecure.NewCredentials()))...)
		if err != nil {
			logger.Error("failed to connect to post-service", "error", err)
			os.Exit(1)
		}
		postClient := postv1.NewPostServiceClient(postConn)

		// gRPC Client setup: Auth Service
		authConn, err := grpc.NewClient(options.AuthService, append(observability.GRPCClientOptions(), grpc.WithTransportCredentials(insecure.NewCredentials()))...)
		if err != nil {
			logger.Error("failed to connect to auth-service", "error", err)
			os.Exit(1)
		}
		authClient := authv1.NewAuthServiceClient(authConn)

		// Watermill Publisher (using Kafka)
		brokers := strings.Split(options.KafkaBrokers, ",")
		kafkaPub, err := kafka.NewPublisher(
			kafka.PublisherConfig{
				Brokers:   brokers,
				Marshaler: kafka.DefaultMarshaler{},
			},
			watermill.NewStdLogger(false, false),
		)
		if err != nil {
			logger.Error("failed to create watermill publisher", "error", err)
			os.Exit(1)
		}
		// Wrap with OTel (Shared)
		publisher := observability.NewTracingPublisher(kafkaPub)

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
