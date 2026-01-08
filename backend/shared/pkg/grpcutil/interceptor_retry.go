package grpcutil

import (
	"context"
	"time"

	"github.com/username/progetto/shared/pkg/resiliency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryOptions configures the retry behavior.
type RetryOptions struct {
	MaxRetries     uint
	BackoffBase    time.Duration
	BackoffMax     time.Duration
	RetriableCodes []codes.Code
}

// DefaultRetryOptions returns reasonable defaults.
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxRetries:  3,
		BackoffBase: 100 * time.Millisecond,
		BackoffMax:  2 * time.Second,
		RetriableCodes: []codes.Code{
			codes.Unavailable,
			codes.ResourceExhausted,
		},
	}
}

// SmartRetryUnaryClientInterceptor returns a new unary client interceptor that retries
// failed RPCs with exponential backoff.
func SmartRetryUnaryClientInterceptor(opts RetryOptions) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, optsCall ...grpc.CallOption) error {
		var err error
		for attempt := uint(0); attempt <= opts.MaxRetries; attempt++ {
			if attempt > 0 {
				backoff := calculateBackoff(attempt, opts.BackoffBase, opts.BackoffMax)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
				}
			}

			err = invoker(ctx, method, req, reply, cc, optsCall...)
			if err == nil {
				return nil
			}

			if !isRetriable(err, opts.RetriableCodes) {
				return err
			}
		}
		return err
	}
}

func calculateBackoff(attempt uint, base, max time.Duration) time.Duration {
	backoff := base * (1 << (attempt - 1))
	if backoff > max {
		return max
	}
	return backoff
}

func isRetriable(err error, codesList []codes.Code) bool {
	// 1. Check if it is a PermanentError (User priority)
	if resiliency.IsPermanentError(err) {
		return false
	}

	// 2. Check gRPC Status Code
	s, ok := status.FromError(err)
	if !ok {
		// Non-gRPC error? Assume not retriable unless we want to retry connection errors.
		// Usually gRPC returns connection errors as Unavailable.
		return false
	}

	for _, code := range codesList {
		if s.Code() == code {
			return true
		}
	}

	return false
}
