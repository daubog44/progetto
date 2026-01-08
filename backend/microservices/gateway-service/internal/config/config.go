package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                 int
	PostService          string
	AuthService          string
	SearchService        string
	KafkaBrokers         string
	RedisAddr            string
	JWTSecret            string
	OtelServiceName      string
	OtelExporterEndpoint string
}

func Load() *Config {
	cfg := &Config{
		Port:            8888, // Default
		OtelServiceName: "gateway-service",
	}

	if val := os.Getenv("PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			cfg.Port = p
		}
	}
	if envPost := os.Getenv("POST_SERVICE"); envPost != "" {
		cfg.PostService = envPost
	}
	if envAuth := os.Getenv("AUTH_SERVICE"); envAuth != "" {
		cfg.AuthService = envAuth
	}
	if envSearch := os.Getenv("SEARCH_SERVICE"); envSearch != "" {
		cfg.SearchService = envSearch
	}
	if envKafka := os.Getenv("KAFKA_BROKERS"); envKafka != "" {
		cfg.KafkaBrokers = envKafka
	}
	if envRedis := os.Getenv("APP_REDIS_ADDR"); envRedis != "" {
		cfg.RedisAddr = envRedis
	}
	if envJWT := os.Getenv("APP_JWT_SECRET"); envJWT != "" {
		cfg.JWTSecret = envJWT
	}
	if val := os.Getenv("OTEL_SERVICE_NAME"); val != "" {
		cfg.OtelServiceName = val
	}
	if val := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); val != "" {
		cfg.OtelExporterEndpoint = val
	}

	return cfg
}
