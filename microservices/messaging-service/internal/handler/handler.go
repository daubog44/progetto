package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/messaging-service/internal/repository"
)

type Client struct {
	Repo      repository.UserRepository
	Publisher message.Publisher
}

type UserCreatedEvent struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (c *Client) HandleUserCreated(msg *message.Message) error {
	// User Created Event Payload
	var event struct {
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		slog.ErrorContext(msg.Context(), "failed to unmarshal event", "error", err)
		return nil // Ack to avoid infinite loop on bad data
	}

	slog.InfoContext(msg.Context(), "Received user_created event", "user_id", event.UserID)

	// Save to Cassandra with Retry (Idempotent)
	// We use the shared Retry Middleware on the router, so if this fails, we return error and let Watermill retry.
	// BUT, for the "Saga Compensation", if we fail permanently, we might want to publish a failure event.
	// The standard way with Watermill Middleware is: if handler returns error, it retries. If max retries reached, it goes to DLQ or Drops.
	// It doesn't automatically publish a "Compensation Event".
	// So we might want to keep the "internal retry" or use a sophisticated error handler.
	// For simplicity in this step: We rely on middleware for transient errors.
	if err := c.Repo.SaveUser(msg.Context(), event.UserID, event.Email, event.Username); err != nil {
		slog.ErrorContext(msg.Context(), "failed to save user to cassandra", "error", err, "user_id", event.UserID)
		return err // Let middleware retry
	}

	slog.InfoContext(msg.Context(), "Successfully saved user to cassandra", "user_id", event.UserID)
	return nil
}
