package config

import (
	"fmt"
	"os"
)

type Config struct {
	DbDSN                string
	RedisAddr            string
	KafkaBrokers         string
	JwtSecret            string
	OtelExporterEndpoint string
	OtelServiceName      string
}

func Load() *Config {
	return &Config{
		DbDSN:                mustGetEnv("APP_DB_DSN"),
		RedisAddr:            mustGetEnv("APP_REDIS_ADDR"),
		KafkaBrokers:         mustGetEnv("APP_KAFKA_BROKERS"),
		JwtSecret:            mustGetEnv("APP_JWT_SECRET"),
		OtelExporterEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "alloy:4317"),
		OtelServiceName:      getEnv("OTEL_SERVICE_NAME", "auth-service"),
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
