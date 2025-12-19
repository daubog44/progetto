package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mongoURI := os.Getenv("APP_MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}
	logger.Info("starting mongo-service", "uri", mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		logger.Error("failed to connect to mongodb", "error", err)
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("failed to ping mongodb", "error", err)
		os.Exit(1)
	}

	logger.Info("connected to MongoDB successfully")
}
