package config

import (
	"github.com/username/progetto/shared/pkg/config"
)

type Config struct {
	MongoURI             string
	KafkaBrokers         string
	OtelServiceName      string
	OtelExporterEndpoint string
}

func Load() *Config {
	return &Config{
		MongoURI:             config.MustGetEnv("APP_MONGO_URI"),
		KafkaBrokers:         config.MustGetEnv("APP_KAFKA_BROKERS"),
		OtelServiceName:      config.GetEnv("OTEL_SERVICE_NAME", "post-service"),
		OtelExporterEndpoint: config.GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "alloy:4317"),
	}
}
