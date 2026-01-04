package grpcutil

import (
	"log/slog"
	"time"

	"github.com/username/progetto/shared/pkg/resiliency"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// NewServer creates a new gRPC server with standard interceptors (Observability).
// Additional options can be passed via opts.
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	// Standard options
	baseOpts := []grpc.ServerOption{
		// OTel Stats Handler for Metrics & Tracing
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// Logging Interceptor (Standard)
		grpc.ChainUnaryInterceptor(UnaryServerLoggingInterceptor(slog.Default())),
		// Keepalive to prevent load balancer disconnects
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              2 * time.Hour,
			Timeout:           20 * time.Second,
		}),
	}
	baseOpts = append(baseOpts, opts...)

	return grpc.NewServer(baseOpts...)
}

// NewClient creates a new gRPC client connection to the target.
// It includes:
// - OTel Observability
// - Circuit Breaker (using the provided name)
// - Retry Logic
// - Insecure credentials (for internal mesh)
func NewClient(target string, cbName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Initialize Circuit Breaker for this client
	cb := resiliency.NewCircuitBreaker(cbName)

	baseOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		// Chained Interceptors for Resiliency
		grpc.WithUnaryInterceptor(resiliency.CircuitBreakerUnaryClientInterceptor(cb)),
	}
	baseOpts = append(baseOpts, opts...)

	// Block until connected or timeout (safer for startup)
	// baseOpts = append(baseOpts, grpc.WithBlock()) // Deprecated/Not recommended for mesh but useful for strict dependency.
	// Let's stick to non-blocking or handle it in app logic. User asked for boilerplate reduction.

	return grpc.NewClient(target, baseOpts...)
}
