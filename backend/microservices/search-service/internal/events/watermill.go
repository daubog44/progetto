package events

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/search-service/internal/handler"
	"github.com/username/progetto/shared/pkg/watermillutil"
)

type WatermillManager struct {
	Publisher  message.Publisher
	Subscriber message.Subscriber
	Router     *message.Router
}

// NewRouter creates and configures the Watermill Router.
// Publisher and Subscriber must be passed in (created by App) to handle dependency injection order (Handler needs Publisher).
func NewRouter(logger *slog.Logger, pub message.Publisher, sub message.Subscriber, h *handler.NotificationHandler) (*WatermillManager, error) {

	// Watermill Router
	router, err := watermillutil.NewRouter(logger, watermillutil.RouterOptions{
		CBName:      "search-service",
		PoisonTopic: "dead_letters",
		Publisher:   pub,
		
	})
	if err != nil {
		return nil, err
	}

	// Add Handlers
	router.AddConsumerHandler(
		"search_user_created",
		"user_created", // Topic
		sub,
		h.HandleUserCreated,
	)

	return &WatermillManager{
		Publisher:  pub,
		Subscriber: sub,
		Router:     router,
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
