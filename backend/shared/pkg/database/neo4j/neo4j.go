package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	neo4jtracing "github.com/raito-io/neo4j-tracing"
	"go.opentelemetry.io/otel"
)

// NewNeo4j creates a new Neo4j driver with OpenTelemetry tracing.
func NewNeo4j(ctx context.Context, uri, username, password string) (neo4j.DriverWithContext, error) {
	// Create Neo4jTracer using the global tracer provider
	tracer := neo4jtracing.NewNeo4jTracer(
		neo4jtracing.WithTracerProvider(otel.GetTracerProvider()),
	)

	// Create the driver with tracing enabled
	driver, err := tracer.NewDriverWithContext(
		uri,
		neo4j.BasicAuth(username, password, ""),
		func(config *neo4j.Config) {
			config.UserAgent = "social-service"
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}

	// Verify connection
	if err := driver.VerifyConnectivity(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to verify neo4j connectivity: %w", err)
	}

	return driver, nil
}
