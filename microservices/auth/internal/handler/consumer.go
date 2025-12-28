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
	var payload struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		slog.ErrorContext(msg.Context(), "failed to unmarshal user_creation_failed event", "error", err)
		return nil // Ack, bad format
	}

	slog.InfoContext(msg.Context(), "Compensating user creation", "user_id", payload.UserID, "reason", payload.Reason)

	if err := h.authService.CompensateUserCreation(msg.Context(), payload.UserID); err != nil {
		slog.ErrorContext(msg.Context(), "failed to compensate user creation", "error", err, "user_id", payload.UserID)
		return err // Retry compensation?? yes
	}

	slog.InfoContext(msg.Context(), "Successfully compensated user creation (user deleted)", "user_id", payload.UserID)
	return nil
}
