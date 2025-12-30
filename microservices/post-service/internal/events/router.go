package events

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/post-service/internal/handler"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type EventRouter struct {
	Router     *message.Router
	Subscriber message.Subscriber
	Publisher  message.Publisher
}

func NewEventRouter(logger *slog.Logger, brokers string, publisher message.Publisher, userHandler *handler.UserHandler) (*EventRouter, error) {
	// 1. Subscriber
	subscriber, err := watermillutil.NewKafkaSubscriber(brokers, "post_service_user_sync", logger)
	if err != nil {
		return nil, err
	}

	// 2. Router
	router, err := watermillutil.NewRouter(logger, watermillutil.RouterOptions{
		CBName:      "post-user-sync",
		PoisonTopic: "dead_letters",
		Publisher:   publisher,
		SagaRoutes: map[string]watermillutil.SagaFailureHandler{
			"user_created": userHandler.HandleFailure,
		},
	})
	if err != nil {
		subscriber.Close()
		return nil, err
	}

	// 4. Handler
	router.AddConsumerHandler(
		"post_user_created_handler",
		"user_created",
		subscriber,
		userHandler.HandleCreated,
	)

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
