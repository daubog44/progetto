package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"go.opentelemetry.io/contrib/instrumentation/runtime"

	// Profiling
	"github.com/grafana/pyroscope-go"
)

// Init initializes the full observability stack (Tracing, Logging, Profiling).
// It returns a shutdown function that should be called when the service exits.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	// 1. Initialize Logger (slog with OTel correlation)
	initLogger()

	// 2. Initialize Tracer (OTLP)
	shutdownTracer, err := initTracer(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init tracer: %w", err)
	}

	// 3. Initialize Meter (Prometheus)
	shutdownMeter, err := initMeter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init meter: %w", err)
	}

	// 3.1 Initialize Runtime Metrics
	if err := runtime.Start(); err != nil {
		slog.Error("failed to start runtime metrics", "error", err)
	}

	// 4. Profiling (Pyroscope) - Optional
	if cfg.PyroscopeAddress != "" {
		if err := initPyroscope(cfg); err != nil {
			slog.Error("Failed to init pyroscope", "error", err)
			// specific error here is not blocking the whole app
		}
	}

	// Return aggregate shutdown function
	return func(shutdownCtx context.Context) error {
		var errs []error
		if err := shutdownTracer(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer: %w", err))
		}
		if err := shutdownMeter(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown meter: %w", err))
		}

		if len(errs) > 0 {
			return fmt.Errorf("shutdown errors: %v", errs)
		}
		return nil
	}, nil
}

func initLogger() {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	handler := ContextHandler{Handler: baseHandler}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func initTracer(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	// Configures the exporter to send traces to Alloy via gRPC
	endpoint := cfg.OTLPEndpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		// Fallback: if OTLP fails, maybe log it but don't crash?
		// For now, let's error out to ensure we fix configuration.
		return nil, err
	}

	bsp := trace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider.Shutdown, nil
}

func initPyroscope(cfg Config) error {
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: cfg.ServiceName,
		ServerAddress:   cfg.PyroscopeAddress,
		Logger:          pyroscope.StandardLogger,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
	})
	return err
}
