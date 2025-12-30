package events

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/gateway-service/internal/sse"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type EventRouter struct {
	Router     *message.Router
	Subscriber message.Subscriber
	Publisher  message.Publisher
}

func NewEventRouter(logger *slog.Logger, brokers string, sseHandler *sse.Handler) (*EventRouter, error) {
	// 1. Publisher
	publisher, err := watermillutil.NewKafkaPublisher(brokers, logger)
	if err != nil {
		return nil, err
	}

	// 2. Subscriber
	subscriber, err := watermillutil.NewKafkaSubscriber(brokers, "gateway-service", logger)
	if err != nil {
		publisher.Close()
		return nil, err
	}

	// 3. Router
	router, err := watermillutil.NewRouter(logger, watermillutil.RouterOptions{
		CBName:      "gateway-events",
		PoisonTopic: "dead_letters",
		Publisher:   publisher,
	})
	if err != nil {
		publisher.Close()
		subscriber.Close()
		return nil, err
	}

	// Note: Direct Kafka-to-SSE handlers removed.
	// SSE is now handled by notification-service -> Redis Pub/Sub -> sse.Handler background subscriber.

	return &EventRouter{
		Router:     router,
		Subscriber: subscriber,
		Publisher:  publisher,
	}, nil
}

func (e *EventRouter) Run(ctx context.Context) error {
	return e.Router.Run(ctx)
}

func (e *EventRouter) Close() {
	if e.Publisher != nil {
		e.Publisher.Close()
	}
	if e.Subscriber != nil {
		e.Subscriber.Close()
	}
	if e.Router != nil {
		e.Router.Close()
	}
}
