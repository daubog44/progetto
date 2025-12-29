package watermillutil

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/username/progetto/shared/pkg/observability"
	"github.com/username/progetto/shared/pkg/resiliency"
)

var (
	registryOnce   sync.Once
	metricsBuilder metrics.PrometheusMetricsBuilder
)

// InitMetrics initializes the shared Prometheus registry and Watermill MetricsBuilder
// and starts the HTTP server for metrics on the given address.
// It returns a cancel function to stop the server.
// If addr is empty, it defaults to ":9091" or PROMETHEUS_METRICS_PORT env var.
func InitMetrics(addr string) func() {
	var closeServer func()
	registryOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Failed to initialize Watermill metrics", "panic", r)
				panic(r)
			}
		}()

		if addr == "" {
			port := os.Getenv("PROMETHEUS_METRICS_PORT")
			if port == "" {
				slog.Error("PROMETHEUS_METRICS_PORT environment variable is not set")
				panic("PROMETHEUS_METRICS_PORT is required")
			}
			if !strings.HasPrefix(port, ":") {
				port = ":" + port
			}
			addr = port
		}

		// 1. Create custom Registry
		registry := prometheus.NewRegistry()

		// 3. Initialize Watermill Builder with this registry
		// Namespace empty as per fix
		metricsBuilder = metrics.NewPrometheusMetricsBuilder(registry, "", "")

		// 4. Start HTTP Server manually
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			Registry: registry,
		}))

		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		go func() {
			slog.Info("Starting Prometheus metrics server", "addr", addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Prometheus metrics server failed", "error", err)
			}
		}()

		closeServer = func() {
			slog.Info("Shutting down Prometheus metrics server")
			server.Close()
		}

		slog.Info("Watermill metrics initialized successfully", "addr", addr)
	})
	// If InitMetrics is called after lazy init, closeServer might be nil if we handled it differently,
	// but here we enforce single init.
	// If called twice, registryOnce skips, closeServer is nil.
	if closeServer == nil {
		return func() {}
	}
	return closeServer
}

// initMetricsIfNeeded ensures metrics are initialized.
// It reads configuration from environment variables if not already initialized.
func initMetricsIfNeeded() {
	// If already initialized (check variable or Once), skip.
	// registryOnce.Do handles the check.
	// Pass empty string to trigger env var lookup.
	InitMetrics("")
}

// NewKafkaPublisher creates a Tracing & Metrics instrumented Kafka Publisher.
func NewKafkaPublisher(brokers string, logger *slog.Logger) (message.Publisher, error) {
	wLogger := observability.NewSlogWatermillAdapter(logger)
	initMetricsIfNeeded()

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

	// 1. Decorate with Metrics
	publisherWithMetrics, err := metricsBuilder.DecoratePublisher(publisher)
	if err != nil {
		return nil, err
	}

	// 2. Wrap with Tracing (Order matters: Outer wraps Inner)
	// Tracing(Metrics(Publisher)) -> Tracing spans cover Metrics recording?
	// Usually Tracing should be outermost to capture everything.
	return observability.NewTracingPublisher(publisherWithMetrics), nil
}

// NewKafkaSubscriber creates a Tracing & Metrics instrumented Kafka Subscriber.
func NewKafkaSubscriber(brokers, consumerGroup string, logger *slog.Logger) (message.Subscriber, error) {
	wLogger := observability.NewSlogWatermillAdapter(logger)
	initMetricsIfNeeded()

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

	// 1. Decorate with Metrics
	subscriberWithMetrics, err := metricsBuilder.DecorateSubscriber(subscriber)
	if err != nil {
		return nil, err
	}

	// 2. Wrap with Tracing
	return observability.NewTracingSubscriber(subscriberWithMetrics), nil
}

// RouterOptions defines configuration for the Watermill Router.
type RouterOptions struct {
	// CBName is the Circuit Breaker name. Required.
	CBName string
	// PoisonTopic, if set, enables standard Poison Middleware to this topic.
	PoisonTopic string
	// Publisher is required if PoisonTopic or SagaMiddlewares are used.
	Publisher message.Publisher

	// SagaMiddlewareConfig defines configuration for Saga Poison Middleware.
	// Map of Topic -> SagaFailureHandler
	SagaRoutes map[string]SagaFailureHandler
}

// SagaFailureHandler receives the error and the original message, and returns a compensation message and its topic.
type SagaFailureHandler func(err error, msg *message.Message) (string, *message.Message, error)

// NewRouter creates a Watermill Router with standard middleware (Recovery, Retry, CircuitBreaker) and Metrics.
// New: Supports optional Poison Queue and Saga Compensation Middleware.
func NewRouter(logger *slog.Logger, options RouterOptions) (*message.Router, error) {
	wLogger := observability.NewSlogWatermillAdapter(logger)
	initMetricsIfNeeded()

	router, err := message.NewRouter(message.RouterConfig{}, wLogger)
	if err != nil {
		return nil, err
	}

	// Initialize Circuit Breaker
	if options.CBName == "" {
		options.CBName = "default_cb"
	}
	cb := resiliency.NewCircuitBreaker(options.CBName)

	// Add Router Metrics to the Builder
	metricsBuilder.AddPrometheusRouterMetrics(router)

	// --- 1. Recovery (Bottom/First) ---
	router.AddMiddleware(middleware.Recoverer)

	// --- 2. Poison / Saga Compensation (Before Retry) ---
	// If Retry fails N times, the error bubbles up to Poison/Saga middleware.
	// It catches the error, publishes to DLQ/Failure Topic, and ACKs the original message.

	if options.PoisonTopic != "" && options.Publisher != nil {
		poisonMiddleware, err := middleware.PoisonQueue(
			options.Publisher,
			options.PoisonTopic,
		)
		if err != nil {
			return nil, err
		}
		router.AddMiddleware(poisonMiddleware)
	}

	// Saga Middleware (Custom Poison)
	if len(options.SagaRoutes) > 0 && options.Publisher != nil {
		router.AddMiddleware(SagaPoisonMiddleware(options.Publisher, options.SagaRoutes, logger))
	}

	// --- 3. Retry (After Poison) ---
	// Retries transient errors. If exhausted, it bubbles up.
	router.AddMiddleware(resiliency.DefaultWatermillRetryMiddleware(wLogger).Middleware)

	// --- 4. Circuit Breaker (Top/Last) ---
	// Fails fast before even trying if unhealthy.
	router.AddMiddleware(resiliency.CircuitBreakerMiddleware(cb))

	// Note on Order:
	// Handle(msg) -> CB -> Retry -> Poison -> Handler(msg)
	// If Handler errors:
	// 1. Poison sees error? No, Retry sees error first.
	// 2. Retry retries.
	// 3. If Retry fails N times, it returns error.
	// 4. Poison sees error. Handles it (publishes failure, returns nil).
	// 5. CB sees nil (success).
	// This seems correct.

	// WAIT: Standard Middleware Order in AddMiddleware is executed in order passed?
	// router.AddMiddleware(m1, m2)
	// handling: m1(m2(handler))
	// So msg enters m1 -> m2 -> handler
	// If we want Retry to wrap Handler, and Poison to wrap Retry:
	// Poison(Retry(Handler))
	// So AddMiddleware(Poison, Retry)
	// My Code: Poison (added first in list above?) No...
	//
	// Previous code:
	// router.AddMiddleware(Recoverer, Retry, CB)
	// Recoverer(Retry(CB(Handler)))?
	// Usually Recoverer is outermost generally.
	// Let's verify standard watermill order.
	// "Middlewares are initialized in the order they are passed to AddMiddleware."
	// "So the first middleware passed will be the first one to receive the request."
	//
	// If we want Request -> Poison -> Retry -> Handler
	// 1. Poison receives msg. Calls next(Retry).
	// 2. Retry receives msg. Calls next(Handler).
	// 3. Handler fails.
	// 4. Retry catches, retries. If exhausted, returns error to Poison.
	// 5. Poison catches error. Publishes to DLQ. Returns nil (ack).
	//
	// So correct order is: Poison, Retry.
	// Let's check my logic above:
	// 1. Recoverer (Outermost)
	// 2. Poison / Saga
	// 3. Retry
	// 4. CB
	// Order in AddMiddleware calls:
	// AddMiddleware(Recoverer)
	// AddMiddleware(Poison)
	// AddMiddleware(Retry)
	// AddMiddleware(CB)
	//
	// Result: Recoverer(Poison(Retry(CB(Handler))))
	// This looks correct.

	return router, nil
}

// SagaPoisonMiddleware creates a middleware that handles final failures by publishing a compensation event.
func SagaPoisonMiddleware(publisher message.Publisher, routes map[string]SagaFailureHandler, logger *slog.Logger) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			// Call next handler (Retry middleware should be downstream)
			producedMessages, err := h(msg)
			if err == nil {
				return producedMessages, nil
			}

			// Capture topic
			topic := message.SubscribeTopicFromCtx(msg.Context())
			handlerFunc, ok := routes[topic]
			if !ok {
				// No saga route for this topic, let error bubble (maybe to standard Poison or Ack/Nack)
				return nil, err
			}

			// We have a Saga Handler for this topic's failure
			logger.WarnContext(msg.Context(), "Handling saga failure for topic", "topic", topic, "error", err)

			failTopic, failMsg, compErr := handlerFunc(err, msg)
			if compErr != nil {
				logger.ErrorContext(msg.Context(), "Failed to create compensation message", "error", compErr)
				return nil, err // Bubble up original error
			}

			// Ensure Context is propagated
			if failMsg.Context() == nil {
				failMsg.SetContext(msg.Context())
			}

			// Publish compensation message
			if pubErr := publisher.Publish(failTopic, failMsg); pubErr != nil {
				logger.ErrorContext(msg.Context(), "Failed to publish compensation message", "error", pubErr, "target_topic", failTopic)
				// If we can't publish, we return error so it might be handled by standard mechanism or logged
				return nil, err
			}

			logger.InfoContext(msg.Context(), "Published compensation event", "topic", failTopic)

			// Return nil to ACK original message since we handled the failure via compensation
			return nil, nil
		}
	}
}
