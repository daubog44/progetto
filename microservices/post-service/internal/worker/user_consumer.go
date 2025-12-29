package worker

import (
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/post-service/internal/model"
	"github.com/username/progetto/post-service/internal/repository"
)

type UserConsumer struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

func NewUserConsumer(userRepo repository.UserRepository) *UserConsumer {
	return &UserConsumer{
		userRepo: userRepo,
		logger:   slog.Default().With("component", "user_consumer"),
	}
}

func (c *UserConsumer) Handle(msg *message.Message) error {
	var payload struct {
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		c.logger.ErrorContext(msg.Context(), "failed to unmarshal message", "error", err)
		return nil // Don't retry malformed messages
	}

	c.logger.InfoContext(msg.Context(), "received user_created event", "user_id", payload.UserID)

	user := &model.User{
		ID:       payload.UserID,
		Email:    payload.Email,
		Username: payload.Username,
	}

	if err := c.userRepo.Save(msg.Context(), user); err != nil {
		c.logger.ErrorContext(msg.Context(), "failed to save user", "error", err)
		return err // Retry
	}

	return nil
}

// HandleUserCreationFailure constructs a compensation message for user creation failure.
func (c *UserConsumer) HandleUserCreationFailure(err error, msg *message.Message) (string, *message.Message, error) {
	// 1. Extract UserID from original message (best effort)
	var payload struct {
		UserID string `json:"user_id"`
	}
	_ = json.Unmarshal(msg.Payload, &payload) // Ignore error, if we can't parse, user_id might be empty

	// 2. Create Failure Payload
	failurePayload := struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
		Source string `json:"source"`
	}{
		UserID: payload.UserID,
		Reason: err.Error(),
		Source: "post-service",
	}

	bytes, marshalErr := json.Marshal(failurePayload)
	if marshalErr != nil {
		return "", nil, marshalErr
	}

	// 3. Create Message
	failMsg := message.NewMessage(watermill.NewUUID(), bytes)

	// 4. Return Topic and Message
	return "user_creation_failed", failMsg, nil
}
