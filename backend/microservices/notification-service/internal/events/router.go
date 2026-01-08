package events

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/notification-service/internal/handler"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type EventRouter struct {
	Router     *message.Router
	Subscriber message.Subscriber
	Publisher  message.Publisher
}

func NewEventRouter(logger *slog.Logger, brokers string, publisher message.Publisher, notifHandler *handler.NotificationHandler, userHandler *handler.UserHandler) (*EventRouter, error) {
	subscriber, err := watermillutil.NewKafkaSubscriber(brokers, "notification-service", logger)
	if err != nil {
		return nil, err
	}

	router, err := watermillutil.NewRouter(logger, watermillutil.RouterOptions{
		CBName:      "notification-events",
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

	// Topic handlers
	router.AddConsumerHandler(
		"notifications_router",
		"user_creation_failed",
		subscriber,
		notifHandler.HandleUserCreationFailure,
	)

	// Aggregator handlers
	router.AddConsumerHandler(
		"aggregator_user_created",
		"user_created",
		subscriber,
		userHandler.HandleUserCreated,
	)

	syncTopics := []string{"user_synced_post", "user_synced_social", "user_synced_search"}
	for _, topic := range syncTopics {
		router.AddConsumerHandler(
			fmt.Sprintf("aggregator_%s", topic),
			topic,
			subscriber,
			userHandler.HandleUserSynced,
		)
	}

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
	if e.Subscriber != nil {
		e.Subscriber.Close()
	}
	if e.Router != nil {
		e.Router.Close()
	}
}
