package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load .env
	if err := godotenv.Load(); err != nil {
		logger.Info("no .env file found")
	}

	// Config
	var k = koanf.New(".")
	k.Load(env.Provider("APP_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil)

	mongoURI := k.String("mongo.uri")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

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
