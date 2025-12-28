package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gocql/gocql"
)

type UserRepository interface {
	SaveUser(ctx context.Context, userID, email, username string) error
	Close()
}

type cassandraRepository struct {
	session *gocql.Session
}

// CassandraConfig holds configuration for the driver
type CassandraConfig struct {
	Host           string
	Consistency    string // e.g. "QUORUM", "ONE", "ALL"
	ConnectTimeout time.Duration
	MaxOpenConns   int // NumConns: number of connections per host
}

func parseConsistency(c string) gocql.Consistency {
	switch c {
	case "ANY":
		return gocql.Any
	case "ONE":
		return gocql.One
	case "TWO":
		return gocql.Two
	case "THREE":
		return gocql.Three
	case "QUORUM":
		return gocql.Quorum
	case "ALL":
		return gocql.All
	case "LOCAL_QUORUM":
		return gocql.LocalQuorum
	case "EACH_QUORUM":
		return gocql.EachQuorum
	case "LOCAL_ONE":
		return gocql.LocalOne
	default:
		slog.Warn("Unknown consistency level, defaulting to QUORUM", "level", c)
		return gocql.Quorum
	}
}

func NewCassandraRepository(cfg CassandraConfig) (UserRepository, error) {
	cluster := gocql.NewCluster(cfg.Host)
	cluster.Keyspace = "messaging"
	cluster.Consistency = parseConsistency(cfg.Consistency)
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = cfg.ConnectTimeout

	// Connection Pooling Settings in Gocql
	// NumConns is the number of connections per host. The driver cycles through them.
	// Default is 2. For high throughput, increasing this helps.
	if cfg.MaxOpenConns > 0 {
		cluster.NumConns = cfg.MaxOpenConns
	} else {
		cluster.NumConns = 2 // Default explicitly
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &cassandraRepository{session: session}, nil
}

func (r *cassandraRepository) SaveUser(ctx context.Context, userID, email, username string) error {
	// Log with context for tracing
	slog.InfoContext(ctx, "Saving user to Cassandra", "user_id", userID, "email", email)

	if err := r.session.Query(`INSERT INTO messaging.users (user_id, email, username, created_at) VALUES (?, ?, ?, ?)`,
		userID, email, username, time.Now()).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *cassandraRepository) Close() {
	if r.session != nil {
		r.session.Close()
	}
}
