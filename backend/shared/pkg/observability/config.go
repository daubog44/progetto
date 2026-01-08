package observability

import (
	"os"
)

// Config holds configuration for the observability stack.
// It is designed to be populated from environment variables.
type Config struct {
	ServiceName    string
	ServiceVersion string

	// OTLP Endpoint for Traces and Metrics (e.g., "http://alloy:4317")
	OTLPEndpoint string

	// Pyroscope Address (e.g., "http://pyroscope:4040")
	PyroscopeAddress string
}

// LoadConfigFromEnv loads configuration from standard environment variables.
//
// Expected Env Vars:
// - OTEL_SERVICE_NAME
// - SERVICE_VERSION (optional, defaults to "0.0.0")
// - OTEL_EXPORTER_OTLP_ENDPOINT (defaults to "http://alloy:4317")
// - PYROSCOPE_SERVER_ADDRESS (optional, if empty profiling is disabled)

func LoadConfigFromEnv() Config {
	cfg := Config{
		ServiceName:      os.Getenv("OTEL_SERVICE_NAME"),
		ServiceVersion:   os.Getenv("SERVICE_VERSION"),
		OTLPEndpoint:     os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		PyroscopeAddress: os.Getenv("PYROSCOPE_SERVER_ADDRESS"),
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "unknown-service"
	}
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "0.0.0"
	}
	if cfg.OTLPEndpoint == "" {
		cfg.OTLPEndpoint = "http://alloy:4317"
	}

	return cfg
}
