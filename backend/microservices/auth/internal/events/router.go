package events

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/auth/internal/handler"
	"github.com/username/progetto/auth/internal/service"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type EventRouter struct {
	Router     *message.Router
	Subscriber message.Subscriber
	Publisher  message.Publisher
}

func NewEventRouter(logger *slog.Logger, brokers string, authSvc service.AuthService) (*EventRouter, error) {
	// 1. Publisher (Needed for Poison Queue)
	publisher, err := watermillutil.NewKafkaPublisher(brokers, logger)
	if err != nil {
		return nil, err
	}

	// 2. Subscriber
	subscriber, err := watermillutil.NewKafkaSubscriber(brokers, "auth-service-saga", logger)
	if err != nil {
		publisher.Close()
		return nil, err
	}

	// 3. Router
	router, err := watermillutil.NewRouter(logger, watermillutil.RouterOptions{
		CBName:      "auth-saga-consumer",
		PoisonTopic: "dead_letters",
		Publisher:   publisher,
	})
	if err != nil {
		subscriber.Close()
		publisher.Close()
		return nil, err
	}

	// 4. Handlers
	sagaHandler := handler.NewSagaHandler(authSvc)
	router.AddConsumerHandler(
		"auth_user_creation_failed_handler",
		"user_creation_failed",
		subscriber,
		sagaHandler.HandleUserCreationFailed,
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
