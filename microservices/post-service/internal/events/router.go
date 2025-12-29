package events

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/post-service/internal/worker"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type EventRouter struct {
	Router     *message.Router
	Subscriber message.Subscriber
	Publisher  message.Publisher
}

func NewEventRouter(logger *slog.Logger, brokers string, userConsumer *worker.UserConsumer) (*EventRouter, error) {
	// 1. Publisher (Needed for Poison/Saga)
	publisher, err := watermillutil.NewKafkaPublisher(brokers, logger)
	if err != nil {
		return nil, err
	}

	// 2. Subscriber
	subscriber, err := watermillutil.NewKafkaSubscriber(brokers, "post_service_user_sync", logger)
	if err != nil {
		publisher.Close()
		return nil, err
	}

	// 3. Router
	router, err := watermillutil.NewRouter(logger, watermillutil.RouterOptions{
		CBName:      "post-user-sync",
		PoisonTopic: "dead_letters",
		Publisher:   publisher,
		SagaRoutes: map[string]watermillutil.SagaFailureHandler{
			"user_created": userConsumer.HandleUserCreationFailure,
		},
	})
	if err != nil {
		subscriber.Close()
		publisher.Close()
		return nil, err
	}

	// 4. Handler
	// Using Router middleware (Retry, Recovery) is better than manual loop.
	// Adapt UserConsumer.Handle(msg) to Router HandlerFunc.
	router.AddConsumerHandler(
		"post_user_created_handler",
		"user_created",
		subscriber,
		userConsumer.Handle,
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
