package events

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/messaging-service/internal/handler"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type EventRouter struct {
	Router     *message.Router
	Subscriber message.Subscriber
}

func NewEventRouter(logger *slog.Logger, brokers string, handler *handler.Handler) (*EventRouter, error) {
	// 1. Subscriber
	subscriber, err := watermillutil.NewKafkaSubscriber(brokers, "messaging-service", logger)
	if err != nil {
		return nil, err
	}

	// 2. Router
	// CHE FIGATA
	router, err := watermillutil.NewRouter(logger, watermillutil.RouterOptions{
		CBName:      "messaging-consumer",
		PoisonTopic: "dead_letters",
		Publisher:   handler.Publisher,
	})
	if err != nil {
		subscriber.Close()
		return nil, err
	}

	// 3. Handlers

	return &EventRouter{
		Router:     router,
		Subscriber: subscriber,
	}, nil
}

func (e *EventRouter) Run(ctx context.Context) error {
	return e.Router.Run(ctx)
}

func (e *EventRouter) Close() {
	if e.Subscriber != nil {
		e.Subscriber.Close()
	}
	if e.Router != nil {
		e.Router.Close()
	}
}
