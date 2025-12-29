package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/messaging-service/internal/repository"
)

type Client struct {
	Repo      repository.UserRepository
	Publisher message.Publisher
	Logger    *slog.Logger
}

type UserCreatedEvent struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// UserCreatedPayload is the structure for the user created event payload.
type UserCreatedPayload struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (c *Client) HandleUserCreated(msg *message.Message) error {
	var payload UserCreatedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		c.Logger.ErrorContext(msg.Context(), "failed to unmarshal message", "error", err)
		return nil // Ack, don't retry bad data
	}

	c.Logger.InfoContext(msg.Context(), "Received user_created event", "user_id", payload.UserID)

	// Save to Cassandra with Retry (Idempotent)
	// We use the shared Retry Middleware on the router, so if this fails, we return error and let Watermill retry.
	// BUT, for the "Saga Compensation", if we fail permanently, we might want to publish a failure event.
	// The standard way with Watermill Middleware is: if handler returns error, it retries. If max retries reached, it goes to DLQ or Drops.
	// It doesn't automatically publish a "Compensation Event".
	// So we might want to keep the "internal retry" or use a sophisticated error handler.
	// For simplicity in this step: We rely on middleware for transient errors.
	if err := c.Repo.SaveUser(msg.Context(), payload.UserID, payload.Email, payload.Username); err != nil {
		c.Logger.ErrorContext(msg.Context(), "failed to save user to cassandra", "error", err, "user_id", payload.UserID)
		return err // Let middleware retry
	}

	c.Logger.InfoContext(msg.Context(), "Successfully saved user to cassandra", "user_id", payload.UserID)
	return nil
}

// HandleUserCreationFailure constructs a compensation message for user creation failure.
func (c *Client) HandleUserCreationFailure(err error, msg *message.Message) (string, *message.Message, error) {
	// 1. Extract UserID from original message (best effort)
	var payload UserCreatedPayload
	_ = json.Unmarshal(msg.Payload, &payload) // Ignore error, if we can't parse, user_id might be empty

	// 2. Create Failure Payload
	failurePayload := struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
	}{
		UserID: payload.UserID,
		Reason: err.Error(),
	}

	bytes, marshalErr := json.Marshal(failurePayload)
	if marshalErr != nil {
		return "", nil, marshalErr
	}

	// 3. Create Message
	failMsg := message.NewMessage(watermill.NewUUID(), bytes)
	// Context is set by Middleware (or we can copy it here just in case, but middleware handles it)

	// 4. Return Topic and Message
	return "user_creation_failed", failMsg, nil
}
