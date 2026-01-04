package handler

import (
	"encoding/json"
	"log/slog"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/messaging-service/internal/repository"
	"github.com/username/progetto/shared/pkg/model"
	"github.com/username/progetto/shared/pkg/resiliency"
)

type Handler struct {
	Repo      repository.UserRepository
	Publisher message.Publisher
	Logger    *slog.Logger
}

// UserCreatedPayload is the structure for the user created event payload.
type UserCreatedPayload struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (h *Handler) HandleUserCreated(msg *message.Message) error {
	var payload UserCreatedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.Logger.ErrorContext(msg.Context(), "failed to unmarshal message", "error", err)
		return nil // Ack, don't retry bad data
	}

	h.Logger.InfoContext(msg.Context(), "Received user_created event", "user_id", payload.UserID)

	// Parse and map to shared model
	userIDUint, err := strconv.Atoi(payload.UserID)
	if err != nil {
		h.Logger.ErrorContext(msg.Context(), "failed to parse user_id", "error", err)

		// If we failed to handle it (e.g. can't extract ID at all), we treat it as Permanent Error
		// This stops retry and sends to Poison Queue (dead_letters)
		// SagaPoisonMiddleware will catch this and publish 'user_creation_failed' using HandleUserCreationFailure
		return resiliency.NewPermanentError(err)
	}

	user := model.User{
		ID:       uint(userIDUint),
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := h.Repo.SaveUser(msg.Context(), user); err != nil {
		h.Logger.ErrorContext(msg.Context(), "failed to save user to cassandra", "error", err, "user_id", payload.UserID)
		return err // Let middleware retry
	}

	// Emit user_synced_messaging event
	syncPayload := struct {
		UserID string `json:"user_id"`
	}{
		UserID: payload.UserID,
	}
	syncBytes, _ := json.Marshal(syncPayload)
	syncMsg := message.NewMessage(watermill.NewUUID(), syncBytes)
	syncMsg.SetContext(msg.Context())

	if err := h.Publisher.Publish("user_synced_messaging", syncMsg); err != nil {
		h.Logger.ErrorContext(msg.Context(), "failed to publish user_synced_messaging event", "error", err)
		return err // Retry
	}

	h.Logger.InfoContext(msg.Context(), "Successfully saved user to cassandra and emitted user_synced_messaging", "user_id", payload.UserID)
	return nil
}

// HandleUserCreationFailure constructs a compensation message for user creation failure.
func (h *Handler) HandleUserCreationFailure(err error, msg *message.Message) (string, *message.Message, error) {
	// 1. Extract UserID from metadata (preferred) or payload (fallback)
	userID := msg.Metadata.Get("user_id")
	if userID == "" {
		var payload UserCreatedPayload
		_ = json.Unmarshal(msg.Payload, &payload)
		userID = payload.UserID
	}

	// 2. Create Failure Payload
	failurePayload := struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
	}{
		UserID: userID,
		Reason: err.Error(),
	}

	bytes, marshalErr := json.Marshal(failurePayload)
	if marshalErr != nil {
		return "", nil, marshalErr
	}

	// 3. Create Message
	failMsg := message.NewMessage(watermill.NewUUID(), bytes)
	failMsg.Metadata.Set("user_id", userID)

	// Context is set by Middleware (or we can copy it here just in case, but middleware handles it)

	// 4. Return Topic and Message
	return "user_creation_failed", failMsg, nil
}
