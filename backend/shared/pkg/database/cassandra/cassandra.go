package cassandra

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gocql/gocql/otelgocql"
)

type Config struct {
	Host           string
	Consistency    string
	ConnectTimeout time.Duration
	Keyspace       string // Optional, if you want to connect to a specific keyspace
}

// NewCassandra creates a new Cassandra session with retry logic.
// While gocql doesn't have a direct official OTel plugin like the others,
// basic connectivity is centralized here.
func NewCassandra(cfg Config, logger *slog.Logger) (*gocql.Session, error) {
	var session *gocql.Session
	var err error

	// Retry loop
	for i := 0; i < 30; i++ {
		cluster := gocql.NewCluster(cfg.Host)
		cluster.Consistency = gocql.ParseConsistency(cfg.Consistency)
		cluster.ConnectTimeout = cfg.ConnectTimeout
		if cfg.Keyspace != "" {
			cluster.Keyspace = cfg.Keyspace
		} else {
			cluster.Keyspace = "progetto" // Default for now, or make param
		}

		// Optimization: TokenAware + RoundRobin
		cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

		// OpenTelemetry Instrumentation
		session, err = otelgocql.NewSessionWithTracing(context.Background(), cluster)
		if err == nil {
			return session, nil
		}

		logger.Info("Waiting for Cassandra...", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to cassandra after retries: %w", err)
}
