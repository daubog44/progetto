package config

import (
	"github.com/username/progetto/shared/pkg/config"
)

type Config struct {
	MeiliHost    string
	MeiliKey     string
	KafkaBrokers string
}

func Load() *Config {
	return &Config{
		MeiliHost:    config.GetEnv("APP_MEILI_HOST", "http://meilisearch:7700"),
		MeiliKey:     config.MustGetEnv("MEILI_MASTER_KEY"),
		KafkaBrokers: config.GetEnv("APP_KAFKA_BROKERS", "kafka:29092"),
	}
}
