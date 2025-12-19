package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	uri := os.Getenv("APP_NEO4J_URI")
	user := os.Getenv("APP_NEO4J_USER")
	pass := os.Getenv("APP_NEO4J_PASSWORD")

	if uri == "" {
		uri = "bolt://localhost:7687"
	}
	if user == "" {
		user = "neo4j"
	}
	if pass == "" {
		pass = "password"
	}
	logger.Info("starting neo4j-service", "uri", uri, "user", user)

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, pass, ""))
	if err != nil {
		logger.Error("failed to create neo4j driver", "error", err)
		os.Exit(1)
	}
	defer driver.Close(context.Background())

	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		logger.Error("failed to verify neo4j connectivity", "error", err)
		os.Exit(1)
	}

	logger.Info("connected to Neo4j successfully")
}
