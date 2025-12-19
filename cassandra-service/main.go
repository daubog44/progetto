package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gocql/gocql"
	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
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

	cassHost := k.String("cassandra.host")
	if cassHost == "" {
		cassHost = "localhost"
	}

	cluster := gocql.NewCluster(cassHost)
	cluster.Consistency = gocql.Quorum
	// Just verify creation
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Error("failed to connect to cassandra", "error", err)
		// Don't exit immediately in dev if DB is not ready,
		// but for a data layer service it might be necessary.
		os.Exit(1)
	}
	defer session.Close()

	logger.Info("connected to Cassandra successfully")
}
