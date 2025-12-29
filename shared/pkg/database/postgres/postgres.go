package postgres

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgres creates a new Postgres connection with OpenTelemetry instrumentation.
// It sets connection pooling settings generally suitable for microservices.
func NewPostgres(dsn string, logger *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Add OpenTelemetry instrumentation
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		logger.Error("failed to use otelgorm plugin", "error", err)
	}

	// Connection Pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Standard pooling config, can be parameterized if needed
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// AutoMigrate is a helper to run auto-migrations.
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}
