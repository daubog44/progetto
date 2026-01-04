package redis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis client with OpenTelemetry tracing and metrics.
func NewRedis(addr string, logger *slog.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Test connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	// Enable tracing
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		logger.Error("failed to instrument redis tracing", "error", err)
	}

	// Enable metrics
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		logger.Error("failed to instrument redis metrics", "error", err)
	}

	return rdb, nil
}
