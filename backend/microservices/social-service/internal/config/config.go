package config

import (
	"fmt"
	"os"
)

type Config struct {
	Neo4jUri             string
	Neo4jUser            string
	Neo4jPassword        string
	KafkaBrokers         string
	OtelExporterEndpoint string
	OtelServiceName      string
}

func Load() *Config {
	return &Config{
		Neo4jUri:             mustGetEnv("APP_NEO4J_URI"),
		Neo4jUser:            mustGetEnv("NEO4J_USER"),
		Neo4jPassword:        mustGetEnv("NEO4J_PASSWORD"),
		KafkaBrokers:         mustGetEnv("APP_KAFKA_BROKERS"),
		OtelExporterEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "alloy:4317"),
		OtelServiceName:      getEnv("OTEL_SERVICE_NAME", "social-service"),
	}
}

func mustGetEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Sprintf("Missing required environment variable: %s", key))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
