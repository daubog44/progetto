package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/username/progetto/notification-service/internal/config"
	"github.com/username/progetto/notification-service/internal/events"
	"github.com/username/progetto/notification-service/internal/handler"
	"github.com/username/progetto/notification-service/internal/presencestore"
	"github.com/username/progetto/shared/pkg/database/redis"
	"github.com/username/progetto/shared/pkg/observability"
)

func main() {
	// 1. Config
	cfg := config.Load()

	// 2. Observability
	obsCfg := observability.LoadConfigFromEnv()
	obsCfg.ServiceName = cfg.OtelServiceName
	obsCfg.OTLPEndpoint = cfg.OtelExporterEndpoint

	shutdown, err := observability.Init(context.Background(), obsCfg)
	if err != nil {
		slog.Error("failed to init observability", "error", err)
	}
	defer func() {
		if shutdown != nil {
			shutdown(context.Background())
		}
	}()

	logger := slog.Default()
	slog.Info("starting notification-service")

	// 3. Redis
	rdb, err := redis.NewRedis(cfg.RedisAddr, logger)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	// 4. Presence Store
	store := presencestore.NewStore(rdb)

	// 5. Wiring
	notifHandler := handler.NewNotificationHandler(rdb, store)

	// 6. Event Router
	router, err := events.NewEventRouter(logger, cfg.KafkaBrokers, notifHandler)
	if err != nil {
		slog.Error("failed to init event router", "error", err)
		os.Exit(1)
	}
	defer router.Close()

	// 6. Metrics Server

	// 7. Run
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := router.Run(ctx); err != nil {
			slog.Error("router failed", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down notification-service")
}
