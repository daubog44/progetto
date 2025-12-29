package repository

import (
	"context"
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

func NewCassandraRepository(session *gocql.Session) UserRepository {
	return &cassandraRepository{session: session}
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
