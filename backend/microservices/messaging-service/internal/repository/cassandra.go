package repository

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/username/progetto/shared/pkg/model"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user model.User) error
	Close()
}

type cassandraRepository struct {
	session *gocql.Session
}

func NewCassandraRepository(session *gocql.Session) UserRepository {
	return &cassandraRepository{session: session}
}

func (r *cassandraRepository) SaveUser(ctx context.Context, user model.User) error {
	// Log with context for tracing
	slog.InfoContext(ctx, "Saving user to Cassandra", "user_id", user.ID, "email", user.Email)

	if err := r.session.Query(`INSERT INTO messaging.users (user_id, email, username, created_at) VALUES (?, ?, ?, ?)`,
		strconv.Itoa(int(user.ID)), user.Email, user.Username, time.Now()).WithContext(ctx).Exec(); err != nil {
		return err
	}
	return nil
}

func (r *cassandraRepository) Close() {
	if r.session != nil {
		r.session.Close()
	}
}
