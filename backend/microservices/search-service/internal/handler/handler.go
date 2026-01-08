package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/search-service/internal/search"
	api_errors "github.com/username/progetto/shared/pkg/resiliency"
)

type NotificationHandler struct {
	Meili     *search.MeiliClient
	Publisher message.Publisher
	Logger    *slog.Logger
}

func NewNotificationHandler(meili *search.MeiliClient, pub message.Publisher) *NotificationHandler {
	return &NotificationHandler{
		Meili:     meili,
		Publisher: pub,
		Logger:    slog.Default().With("component", "search_handler"),
	}
}

func (h *NotificationHandler) HandleUserCreated(msg *message.Message) error {
	var payload struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.Logger.Error("failed to unmarshal payload", "error", err)
		return api_errors.NewPermanentError(err)
	}

	h.Logger.Info("indexing user", "user_id", payload.UserID)

	user := search.User{
		ID:       payload.UserID,
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := h.Meili.IndexUser(msg.Context(), user); err != nil {
		h.Logger.Error("failed to index user", "error", err)
		return err // retry
	}

	syncPayload := struct {
		UserID string `json:"user_id"`
	}{
		UserID: payload.UserID,
	}

	b, err := json.Marshal(syncPayload)
	if err != nil {
		return err
	}

	outMsg := message.NewMessage(watermill.NewUUID(), b)
	outMsg.SetContext(msg.Context())

	if err := h.Publisher.Publish("user_synced_search", outMsg); err != nil {
		return err
	}

	h.Logger.Info("user indexed and sync event published", "user_id", payload.UserID)
	return nil
}

// HandleFailure constructs a compensation message for user creation failure.
func (h *NotificationHandler) HandleFailure(err error, msg *message.Message) (string, *message.Message, error) {
	// Try to get UserID from Metadata first (if passed), otherwise from Payload
	userID := msg.Metadata.Get("user_id")
	if userID == "" {
		var payload struct {
			UserID string `json:"user_id"`
		}
		// Best effort unmarshal
		_ = json.Unmarshal(msg.Payload, &payload)
		userID = payload.UserID
	}

	h.Logger.WarnContext(msg.Context(), "handling failure for user_created", "user_id", userID, "error", err)

	failurePayload := struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
		Source string `json:"source"`
	}{
		UserID: userID,
		Reason: err.Error(),
		Source: "search-service",
	}

	bytes, marshalErr := json.Marshal(failurePayload)
	if marshalErr != nil {
		return "", nil, marshalErr
	}

	failMsg := message.NewMessage(watermill.NewUUID(), bytes)
	failMsg.Metadata.Set("user_id", userID)

	// Return the topic 'user_creation_failed' and the message
	return "user_creation_failed", failMsg, nil
}
