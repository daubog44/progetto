package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppName               string
	AppEnv                string
	Neo4jUri              string
	Neo4jUser             string
	Neo4jPassword         string
	KafkaBrokers          string
	OtelExporterEndpoint  string
	OtelServiceName       string
	PrometheusMetricsPort int
}

func Load() *Config {
	return &Config{
		AppName:               mustGetEnv("APP_NAME"),
		AppEnv:                mustGetEnv("APP_ENV"), // Ensure this is set in docker-compose if strictly required, or keep it optional? User said "manca una env", usually refers to config. Let's make all strict as requested.
		Neo4jUri:              mustGetEnv("APP_NEO4J_URI"),
		Neo4jUser:             mustGetEnv("NEO4J_USER"),
		Neo4jPassword:         mustGetEnv("NEO4J_PASSWORD"),
		KafkaBrokers:          mustGetEnv("APP_KAFKA_BROKERS"),
		OtelExporterEndpoint:  mustGetEnv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OtelServiceName:       mustGetEnv("OTEL_SERVICE_NAME"),
		PrometheusMetricsPort: mustGetEnvAsInt("PROMETHEUS_METRICS_PORT"),
	}
}

func mustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("Missing required environment variable: %s", key))
	}
	return value
}

func mustGetEnvAsInt(key string) int {
	strValue := mustGetEnv(key)
	value, err := strconv.Atoi(strValue)
	if err != nil {
		panic(fmt.Sprintf("Invalid integer for environment variable %s: %v", key, err))
	}
	return value
}
