package events

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/gateway-service/internal/sse"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type EventRouter struct {
	Router     *message.Router
	Subscriber message.Subscriber
	Publisher  message.Publisher // Kept for future use, though currently unused
}

func NewEventRouter(logger *slog.Logger, brokers string, sseHandler *sse.Handler) (*EventRouter, error) {
	// 1. Publisher
	publisher, err := watermillutil.NewKafkaPublisher(brokers, logger)
	if err != nil {
		return nil, err
	}

	// 2. Subscriber
	subscriber, err := watermillutil.NewKafkaSubscriber(brokers, "gateway-service-sse", logger)
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

	// 4. Handlers
	broadcastHandler := func(msg *message.Message) error {
		// user_created payload: {user_id, ...}
		// user_creation_failed payload: {user_id, reason}
		// Common: has user_id
		var payload struct {
			UserID string `json:"user_id"`
		}
		// Best effort unmarshal to get ID
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return nil // skip
		}

		// Topic name as event type
		topic := message.SubscribeTopicFromCtx(msg.Context())

		sseHandler.Broadcast(msg.Context(), payload.UserID, topic, string(msg.Payload))
		return nil
	}

	router.AddConsumerHandler("gateway_user_creation_failed", "user_creation_failed", subscriber, broadcastHandler)

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
