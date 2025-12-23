package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	postv1 "github.com/username/progetto/proto/gen/go/post/v1"
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
	PostService  string `help:"Address of the post service" default:"mongo-service:50051"`
	KafkaBrokers string `help:"Kafka brokers (comma-separated)" default:"kafka:29092"`
}

func main() {
	var api huma.API

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new router & API
		router := chi.NewMux()
		api = humachi.New(router, huma.DefaultConfig("Gateway API", "1.0.0"))

		// gRPC Client setup
		conn, err := grpc.NewClient(options.PostService, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Error("failed to connect to post-service", "error", err)
			os.Exit(1)
		}
		postClient := postv1.NewPostServiceClient(conn)

		// Watermill Publisher (using Kafka)
		brokers := strings.Split(options.KafkaBrokers, ",")
		publisher, err := kafka.NewPublisher(
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

		// Register Post Routes
		RegisterPostRoutes(api, postClient, publisher)

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

		// Tell the CLI how to start your server.
		hooks.OnStart(func() {
			fmt.Printf("Starting gateway on port %d...\n", options.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
		})
		hooks.OnStop(func() {
			conn.Close()
		})
	})

	cli.Root().AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI spec",
		Run: func(cmd *cobra.Command, args []string) {
			b, _ := api.OpenAPI().DowngradeYAML()
			fmt.Println(string(b))
		},
	})

	cli.Run()
}
