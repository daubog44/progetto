package config

import (
	"fmt"
	"os"
)

type Config struct {
	MongoURI             string
	KafkaBrokers         string
	OtelServiceName      string
	OtelExporterEndpoint string
}

func Load() *Config {
	return &Config{
		MongoURI:             mustGetEnv("APP_MONGO_URI"),
		KafkaBrokers:         mustGetEnv("APP_KAFKA_BROKERS"),
		OtelServiceName:      getEnv("OTEL_SERVICE_NAME", "post-service"),
		OtelExporterEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "alloy:4317"),
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
