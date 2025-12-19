package main

import (
	"log/slog"
	"os"

	"github.com/gocql/gocql"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cassHost := os.Getenv("APP_CASSANDRA_HOST")
	if cassHost == "" {
		cassHost = "localhost"
	}
	logger.Info("starting cassandra-service", "host", cassHost)

	cluster := gocql.NewCluster(cassHost)
	cluster.Consistency = gocql.Quorum
	// Just verify creation
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Error("failed to connect to cassandra", "error", err)
		os.Exit(1)
	}
	defer session.Close()

	logger.Info("connected to Cassandra successfully")
}
