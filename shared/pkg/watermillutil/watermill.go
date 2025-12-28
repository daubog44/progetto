package watermillutil

import (
	"log/slog"
	"strings"

	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/resiliency"
)

// NewKafkaPublisher creates a Tracing-instrumented Kafka Publisher.
func NewKafkaPublisher(brokers string, logger *slog.Logger) (message.Publisher, error) {
	wLogger := observability.NewSlogWatermillAdapter(logger)

	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   strings.Split(brokers, ","),
			Marshaler: kafka.DefaultMarshaler{},
		},
		wLogger,
	)
	if err != nil {
		return nil, err
	}

	// Wrap with Tracing
	return observability.NewTracingPublisher(publisher), nil
}

// NewKafkaSubscriber creates a Tracing-instrumented Kafka Subscriber.
func NewKafkaSubscriber(brokers, consumerGroup string, logger *slog.Logger) (message.Subscriber, error) {
	wLogger := observability.NewSlogWatermillAdapter(logger)

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       strings.Split(brokers, ","),
			Unmarshaler:   kafka.DefaultMarshaler{},
			ConsumerGroup: consumerGroup,
		},
		wLogger,
	)
	if err != nil {
		return nil, err
	}

	// Wrap with Tracing
	return observability.NewTracingSubscriber(subscriber), nil
}

// NewRouter creates a Watermill Router with standard middleware (Recovery, Retry, CircuitBreaker).
// cbName identifies the circuit breaker instance for this router's consumers.
func NewRouter(logger *slog.Logger, cbName string) (*message.Router, error) {
	wLogger := observability.NewSlogWatermillAdapter(logger)

	router, err := message.NewRouter(message.RouterConfig{}, wLogger)
	if err != nil {
		return nil, err
	}

	// Initialize Circuit Breaker
	cb := resiliency.NewCircuitBreaker(cbName)

	// Standard Middleware
	router.AddMiddleware(
		// 1. Recovery: Catches panics from all inner handlers
		middleware.Recoverer,
		// 2. Retry: Retries the handler if it fails (and CB doesn't fail fast)
		resiliency.DefaultWatermillRetryMiddleware().Middleware,
		// 3. Circuit Breaker: Fail fast if the service is unhealthy
		resiliency.CircuitBreakerMiddleware(cb),
	)

	return router, nil
}
