package meilisearch

import (
	"net/http"
	"time"

	meilisearch "github.com/meilisearch/meilisearch-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewClient creates a new Meilisearch client with OpenTelemetry instrumentation.
func NewClient(host, key string) meilisearch.ServiceManager {
	// Instrumented HTTP Client
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return "meilisearch-request: " + r.URL.Path
			}),
		),
		Timeout: 5 * time.Second,
	}

	client := meilisearch.New(host,
		meilisearch.WithAPIKey(key),
		meilisearch.WithCustomClient(httpClient),
	)

	return client
}
