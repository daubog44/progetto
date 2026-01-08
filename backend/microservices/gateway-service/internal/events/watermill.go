package events

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type WatermillManager struct {
	Router     *message.Router
	Subscriber message.Subscriber
	Publisher  message.Publisher
}

func NewWatermillManager(logger *slog.Logger, pub message.Publisher, sub message.Subscriber) (*WatermillManager, error) {

	// Watermill Router
	router, err := watermillutil.NewRouter(logger, watermillutil.RouterOptions{
		CBName:      "gateway-events",
		PoisonTopic: "dead_letters",
		Publisher:   pub,
	})
	if err != nil {
		return nil, err
	}

	// Note: Direct Kafka-to-SSE handlers removed.
	// SSE is now handled by notification-service -> Redis Pub/Sub -> sse.Handler background subscriber.

	return &WatermillManager{
		Router:     router,
		Subscriber: sub,
		Publisher:  pub,
	}, nil
}

func (w *WatermillManager) Close() {
	if w.Publisher != nil {
		w.Publisher.Close()
	}
	if w.Subscriber != nil {
		w.Subscriber.Close()
	}
	if w.Router != nil {
		w.Router.Close()
	}
}
