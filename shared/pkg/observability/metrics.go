package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// initMeter initializes the OpenTelemetry MeterProvider with an OTLP exporter.
// It returns a shutdown function.
func initMeter(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	// Configures the exporter to send metrics to Alloy via gRPC
	endpoint := cfg.OTLPEndpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(provider)

	return provider.Shutdown, nil
}

// Middleware returns a replacement for the http handler that tracks metrics and traces.
func Middleware(next http.Handler) http.Handler {
	meter := otel.GetMeterProvider().Meter("github.com/daubog44/progetto/shared/pkg/observability")
	tracer := otel.GetTracerProvider().Tracer("github.com/daubog44/progetto/shared/pkg/observability")

	requestCounter, _ := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)

	requestDuration, _ := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Start a span
		ctx, span := tracer.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPTargetKey.String(r.URL.Path),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Wrap writer to capture status code
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Pass the context with span to the next handlers
		next.ServeHTTP(ww, r.WithContext(ctx))

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", ww.status)

		// Record metrics
		attrs := metric.WithAttributes(
			attribute.String("method", r.Method),
			attribute.String("path", r.URL.Path),
			attribute.String("status", status),
		)
		requestCounter.Add(ctx, 1, attrs)
		requestDuration.Record(ctx, duration, attrs)

		// Update span status and attributes
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(ww.status))
		if ww.status >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", ww.status))

			// Log the error with trace context
			slog.ErrorContext(ctx, "HTTP request failed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", duration*1000,
			)
		} else {
			span.SetStatus(codes.Ok, "")
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
