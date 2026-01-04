package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/spf13/cobra"
	"github.com/username/progetto/gateway-service/internal/api"
	"github.com/username/progetto/shared/pkg/observability"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

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

	cli := api.CLI(logger, httpAPI)
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
