package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"github.com/username/progetto/notification-service/internal/presencestore"
	"github.com/username/progetto/shared/pkg/presence"
)

type NotificationHandler struct {
	Redis    *redis.Client
	Presence *presencestore.Store
	Logger   *slog.Logger
}

func NewNotificationHandler(rdb *redis.Client, store *presencestore.Store) *NotificationHandler {
	return &NotificationHandler{
		Redis:    rdb,
		Presence: store,
		Logger:   slog.Default().With("component", "notification_handler"),
	}
}

func (h *NotificationHandler) HandleUserCreationFailure(msg *message.Message) error {
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.Logger.Error("failed to unmarshal payload", "error", err)
		return nil // skip
	}

	if payload.UserID == "" {
		return nil
	}

	instanceID, err := h.Presence.GetUserGateway(msg.Context(), payload.UserID)
	if err != nil {
		return err
	}

	if instanceID == "" {
		h.Logger.Debug("user not connected, skipping", "user_id", payload.UserID)
		return nil
	}

	// Used explicitly for failure notifications but it is for the frontend
	topic := "user_creation_failed"
	channel := fmt.Sprintf("gateway_events:%s", instanceID)

	sseEvent := presence.TargetedEvent{
		UserID:  payload.UserID,
		Type:    topic,
		Payload: string(msg.Payload),
	}

	b, _ := json.Marshal(sseEvent)
	if err := h.Redis.Publish(msg.Context(), channel, b).Err(); err != nil {
		return err
	}

	h.Logger.Info("routed notification to gateway", "user_id", payload.UserID, "instance_id", instanceID)
	return nil
}
