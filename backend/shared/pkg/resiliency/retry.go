package resiliency

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
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

// SmartRetryMiddleware implements a retry logic that respects PermanentError.
func SmartRetryMiddleware(logger watermill.LoggerAdapter) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			maxRetries := 5
			initialInterval := 50 * time.Millisecond
			maxInterval := 2 * time.Second
			multiplier := 1.5

			currentInterval := initialInterval
			var err error
			var events []*message.Message

			for i := 0; i <= maxRetries; i++ {
				events, err = h(msg)
				if err == nil {
					return events, nil
				}

				// Check if error is Permanent
				if IsPermanentError(err) {
					// Unwrap and return immediately to allow Poison Queue to handle it
					return nil, err
				}

				if i == maxRetries {
					break // Bubbles up final error
				}

				// Log retry
				logger.Error("Error processing message, retrying...", err, map[string]interface{}{
					"retry_no":     i + 1,
					"max_retries":  maxRetries,
					"wait_time":    currentInterval,
					"message_uuid": msg.UUID,
				})

				// Wait
				select {
				case <-time.After(currentInterval):
				case <-msg.Context().Done():
					return nil, msg.Context().Err()
				}

				// Backoff
				currentInterval = time.Duration(float64(currentInterval) * multiplier)
				if currentInterval > maxInterval {
					currentInterval = maxInterval
				}
			}

			return nil, err
		}
	}
}
