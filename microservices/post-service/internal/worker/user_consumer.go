package worker

import (
	"encoding/json"
	"log/slog"

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
		c.logger.Error("failed to unmarshal message", "error", err)
		return nil // Don't retry malformed messages
	}

	c.logger.Info("received user_created event", "user_id", payload.UserID)

	user := &model.User{
		ID:       payload.UserID,
		Email:    payload.Email,
		Username: payload.Username,
	}

	if err := c.userRepo.Save(msg.Context(), user); err != nil {
		c.logger.Error("failed to save user", "error", err)
		return err // Retry
	}

	return nil
}
