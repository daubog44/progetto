package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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

	uri := k.String("neo4j.uri")
	user := k.String("neo4j.user")
	pass := k.String("neo4j.password")

	if uri == "" {
		uri = "bolt://localhost:7687"
	}
	if user == "" {
		user = "neo4j"
	}
	if pass == "" {
		pass = "password"
	}

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
