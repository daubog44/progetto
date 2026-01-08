package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

// NewMongo creates a new MongoDB client and returns a specific database.
// It includes OpenTelemetry monitoring.
func NewMongo(ctx context.Context, uri string, dbName string) (*mongo.Client, *mongo.Database, error) {
	clientOpts := options.Client().ApplyURI(uri)

	// Add OpenTelemetry instrumentation
	clientOpts.Monitor = otelmongo.NewMonitor()

	// Short timeout for initial connection
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Verify connection
	if err := client.Ping(connectCtx, nil); err != nil {
		return nil, nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	db := client.Database(dbName)
	return client, db, nil
}
