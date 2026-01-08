package config

import (
	"fmt"
	"os"
)

type Config struct {
	KafkaBrokers         string
	RedisAddr            string
	OtelExporterEndpoint string
	OtelServiceName      string
}

func Load() *Config {
	return &Config{
		KafkaBrokers:         mustGetEnv("APP_KAFKA_BROKERS"),
		RedisAddr:            mustGetEnv("APP_REDIS_ADDR"),
		OtelExporterEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "alloy:4317"),
		OtelServiceName:      getEnv("OTEL_SERVICE_NAME", "notification-service"),
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
