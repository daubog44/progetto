package handler

import (
	"encoding/json"
	"log/slog"

	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/post-service/internal/model"
	"github.com/username/progetto/post-service/internal/repository"
	"github.com/username/progetto/shared/pkg/resiliency"
)

type UserHandler struct {
	Repo      repository.UserRepository
	Publisher message.Publisher
	Logger    *slog.Logger
}

func NewUserHandler(repo repository.UserRepository, publisher message.Publisher) *UserHandler {
	return &UserHandler{
		Repo:      repo,
		Publisher: publisher,
		Logger:    slog.Default().With("component", "user_handler"),
	}
}

func (h *UserHandler) HandleCreated(msg *message.Message) error {
	var payload struct {
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.Logger.ErrorContext(msg.Context(), "failed to unmarshal message", "error", err)
		return nil // Don't retry malformed messages
	}

	h.Logger.InfoContext(msg.Context(), "received user_created event", "user_id", payload.UserID)

	userIDUint, err := strconv.Atoi(payload.UserID)
	if err != nil {
		h.Logger.ErrorContext(msg.Context(), "failed to parse user_id", "error", err)

		// SagaPoisonMiddleware will catch this and publish 'user_creation_failed' using HandleFailure
		return resiliency.NewPermanentError(err)
	}

	user := &model.User{
		ID:       uint(userIDUint),
		Email:    payload.Email,
		Username: payload.Username,
	}

	if err := h.Repo.Save(msg.Context(), user); err != nil {
		h.Logger.ErrorContext(msg.Context(), "failed to save user", "error", err)
		return err // Retry
	}

	// Emit user_synced_post event
	syncPayload := struct {
		UserID string `json:"user_id"`
	}{
		UserID: payload.UserID,
	}
	syncBytes, _ := json.Marshal(syncPayload)
	syncMsg := message.NewMessage(watermill.NewUUID(), syncBytes)
	syncMsg.SetContext(msg.Context())

	if err := h.Publisher.Publish("user_synced_post", syncMsg); err != nil {
		h.Logger.ErrorContext(msg.Context(), "failed to publish user_synced_post event", "error", err)
		return err // Retry
	}

	h.Logger.InfoContext(msg.Context(), "successfully published user_synced_post event", "user_id", payload.UserID)

	return nil
}

// HandleFailure constructs a compensation message for user creation failure.
func (h *UserHandler) HandleFailure(err error, msg *message.Message) (string, *message.Message, error) {
	userID := msg.Metadata.Get("user_id")
	if userID == "" {
		var payload struct {
			UserID string `json:"user_id"`
		}
		_ = json.Unmarshal(msg.Payload, &payload)
		userID = payload.UserID
	}

	failurePayload := struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
		Source string `json:"source"`
	}{
		UserID: userID,
		Reason: err.Error(),
		Source: "post-service",
	}

	bytes, marshalErr := json.Marshal(failurePayload)
	if marshalErr != nil {
		return "", nil, marshalErr
	}

	failMsg := message.NewMessage(watermill.NewUUID(), bytes)
	failMsg.Metadata.Set("user_id", userID)
	return "user_creation_failed", failMsg, nil
}
