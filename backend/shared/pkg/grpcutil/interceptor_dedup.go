package grpcutil

import (
	"context"
	"time"

	"github.com/username/progetto/shared/pkg/deduplication"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const RequestIDHeader = "x-request-id"

// UnaryServerDeduplicationInterceptor returns a new unary server interceptor that performs request deduplication.
// It looks for 'x-request-id' in the metadata.
// for now not used
func UnaryServerDeduplicationInterceptor(deduplicator deduplication.Deduplicator, ttl time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		// Get Request ID
		ids := md.Get(RequestIDHeader)
		if len(ids) == 0 || ids[0] == "" {
			// No Request ID, skip deduplication
			return handler(ctx, req)
		}
		requestID := ids[0]

		// Check Uniqueness
		unique, err := deduplicator.IsUnique(ctx, requestID, ttl)
		if err != nil {
			// Fail open or closed?
			// Let's fail closed (return error) because system is unstable
			return nil, status.Errorf(codes.Internal, "deduplication check failed: %v", err)
		}

		if !unique {
			return nil, status.Errorf(codes.AlreadyExists, "duplicate request id: %s", requestID)
		}

		return handler(ctx, req)
	}
}
