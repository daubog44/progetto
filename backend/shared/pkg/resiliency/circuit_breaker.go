package resiliency

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
)

// NewCircuitBreaker creates a standard configured gobreaker
func NewCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	var st gobreaker.Settings
	st.Name = name
	st.MaxRequests = 5 // Allow 5 concurrent requests in half-open state
	st.Interval = 10 * time.Second
	st.Timeout = 30 * time.Second // Time to stay open before trying again
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		// Trip if >= 3 requests and >60% failed
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 3 && failureRatio >= 0.6
	}
	st.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		slog.Warn("Circuit Breaker State Change", "name", name, "from", from.String(), "to", to.String())
	}
	return gobreaker.NewCircuitBreaker(st)
}

// GRPC Circuit Breaker Interceptor
func CircuitBreakerUnaryClientInterceptor(cb *gobreaker.CircuitBreaker) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})
		return err
	}
}

// Watermill Middleware
func CircuitBreakerMiddleware(cb *gobreaker.CircuitBreaker) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			res, err := cb.Execute(func() (interface{}, error) {
				return h(msg)
			})

			if err != nil {
				// If circuit breaker is open, gobreaker returns ErrOpenState.
				// For Watermill, if we return error, it might Nack and retry depending on config.
				// If CB is open, we probably want to Nack to retry later (hoping CB closes).
				return nil, err
			}

			// If success, user handler returns slice of messages and possible error
			// If CB Execute succeeds, res is [] *message.Message
			if msgs, ok := res.([]*message.Message); ok {
				return msgs, nil
			}
			return nil, errors.New("unexpected return type from circuit breaker")
		}
	}
}
