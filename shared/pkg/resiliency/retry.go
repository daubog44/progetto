package resiliency

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/cenkalti/backoff/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPC Retry Interceptor
func RetryUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		operation := func() error {
			err := invoker(ctx, method, req, reply, cc, opts...)
			// Only retry appropriate errors
			if err != nil {
				s, ok := status.FromError(err)
				if ok {
					switch s.Code() {
					case codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded:
						return err // Retry these
					}
				}
				// Don't retry other errors (e.g. InvalidArgument, NotFound) using Backoff.Permanent
				return backoff.Permanent(err)
			}
			return nil
		}

		// Configurable backoff
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = 50 * time.Millisecond
		b.MaxInterval = 2 * time.Second
		b.MaxElapsedTime = 5 * time.Second

		return backoff.Retry(operation, b)
	}
}

// Watermill Retry Middleware Configuration
// Watermill has a built-in middleware, we can just provide a standard config helper
func DefaultWatermillRetryMiddleware(logger watermill.LoggerAdapter) middleware.Retry {
	return middleware.Retry{
		MaxRetries:      5,
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     2 * time.Second,
		Multiplier:      1.5,
		Logger:          logger,
	}
}
