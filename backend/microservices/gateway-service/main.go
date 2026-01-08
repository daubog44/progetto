package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/username/progetto/gateway-service/internal/app"
	"github.com/username/progetto/gateway-service/internal/config"
	"github.com/username/progetto/shared/pkg/observability"
)

func main() {
	// Initialize Config
	cfg := config.Load()

	// Initialize Observability
	obsCfg := observability.LoadConfigFromEnv()
	// Use config overrides if available (loaded from env in config anyway)
	obsCfg.ServiceName = cfg.OtelServiceName
	obsCfg.OTLPEndpoint = cfg.OtelExporterEndpoint

	shutdown, err := observability.Init(context.Background(), obsCfg)
	if err != nil {
		slog.Error("failed to init observability", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown observability", "error", err)
		}
	}()

	logger := slog.Default()
	slog.Info("starting gateway-service...", "config", cfg)

	// Initialize Application
	application, err := app.New(cfg, logger)
	if err != nil {
		slog.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	// Run Application
	if err := application.Run(); err != nil {
		slog.Error("application run error", "error", err)
		os.Exit(1)
	}
}
