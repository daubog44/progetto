package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/username/progetto/search-service/internal/app"
	"github.com/username/progetto/search-service/internal/config"
	"github.com/username/progetto/shared/pkg/observability"
)

func main() {
	// Initialize Observability
	cfgObs := observability.LoadConfigFromEnv()
	shutdown, err := observability.Init(context.Background(), cfgObs)
	if err != nil {
		slog.Error("failed to init observability", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown observability", "error", err)
		}
	}()

	// Load Configuration
	cfg := config.Load()
	slog.Info("Starting search-service...", "config", cfg)

	// Initialize Application
	application, err := app.New(cfg)
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
