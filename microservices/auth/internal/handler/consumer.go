package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/auth/internal/service"
)

type SagaHandler struct {
	authService service.AuthService
}

func NewSagaHandler(authService service.AuthService) *SagaHandler {
	return &SagaHandler{authService: authService}
}

func (h *SagaHandler) HandleUserCreationFailed(msg *message.Message) error {
	// 1. Try to unmarshal payload
	var payload struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
	}

	err := json.Unmarshal(msg.Payload, &payload)
	userID := payload.UserID

	// 2. Fallback to metadata if payload failed or ID is empty
	if err != nil || userID == "" {
		userID = msg.Metadata.Get("user_id")
		if userID == "" {
			slog.ErrorContext(msg.Context(), "failed to extract user_id from failure event (payload and metadata missing)", "error", err)
			return nil // Ack, nothing we can do
		}
		// If we found ID in metadata, we can proceed even if JSON was bad
	}

	slog.InfoContext(msg.Context(), "Compensating user creation", "user_id", userID, "reason", payload.Reason)

	if err := h.authService.CompensateUserCreation(msg.Context(), userID); err != nil {
		slog.ErrorContext(msg.Context(), "failed to compensate user creation", "error", err, "user_id", userID)
		return err // Retry compensation?? yes
	}

	slog.InfoContext(msg.Context(), "Successfully compensated user creation (user deleted)", "user_id", userID)
	return nil
}
