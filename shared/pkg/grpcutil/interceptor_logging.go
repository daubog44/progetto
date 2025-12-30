package grpcutil

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerLoggingInterceptor returns a new unary server interceptor that logs request start and end.
func UnaryServerLoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Log Start
		logger.InfoContext(ctx, "gRPC Call started", "method", info.FullMethod)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := status.Code(err)

		// Log End
		if err != nil {
			logger.ErrorContext(ctx, "gRPC Call completed with error",
				"method", info.FullMethod,
				"duration", duration,
				"code", code.String(),
				"error", err,
			)
		} else {
			logger.InfoContext(ctx, "gRPC Call completed",
				"method", info.FullMethod,
				"duration", duration,
				"code", code.String(),
			)
		}

		return resp, err
	}
}
