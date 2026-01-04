package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"github.com/username/progetto/notification-service/internal/presencestore"
	"github.com/username/progetto/shared/pkg/presence"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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

func (h *NotificationHandler) HandleNotification(msg *message.Message) error {
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

	topic := message.SubscribeTopicFromCtx(msg.Context())
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

func (h *NotificationHandler) HandleUserCreated(msg *message.Message) error {
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil
	}

	key := fmt.Sprintf("registration:%s:pending_syncs", payload.UserID)
	pipe := h.Redis.Pipeline()
	pipe.SAdd(msg.Context(), key, "post", "social", "messaging")
	pipe.Expire(msg.Context(), key, 1*time.Hour)

	if _, err := pipe.Exec(msg.Context()); err != nil {
		return err
	}

	h.Logger.Info("initialized tracking for user registration", "user_id", payload.UserID)
	return nil
}

func (h *NotificationHandler) HandleUserSynced(msg *message.Message) error {
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil
	}

	topic := message.SubscribeTopicFromCtx(msg.Context())
	service := ""
	switch topic {
	case "user_synced_post":
		service = "post"
	case "user_synced_social":
		service = "social"
	case "user_synced_messaging":
		service = "messaging"
	}

	if service == "" {
		return nil
	}

	key := fmt.Sprintf("registration:%s:pending_syncs", payload.UserID)
	if err := h.Redis.SRem(msg.Context(), key, service).Err(); err != nil {
		return err
	}

	count, err := h.Redis.SCard(msg.Context(), key).Result()
	if err != nil {
		return err
	}

	h.Logger.Info("received sync event", "user_id", payload.UserID, "service", service, "remaining", count)

	if count == 0 {
		return h.triggerOnboardingCompleted(msg.Context(), payload.UserID)
	}

	return nil
}

func (h *NotificationHandler) triggerOnboardingCompleted(ctx context.Context, userID string) error {
	h.Logger.Info("onboarding completed for user", "user_id", userID)

	instanceID, err := h.Presence.GetUserGateway(ctx, userID)
	if err != nil {
		return err
	}

	if instanceID == "" {
		h.Logger.Debug("user disconnected before onboarding completed", "user_id", userID)
		return nil
	}

	channel := fmt.Sprintf("gateway_events:%s", instanceID)

	// Inject Trace Context
	traceContext := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(traceContext))

	completionEvent := presence.TargetedEvent{
		UserID:       userID,
		Type:         "onboarding_completed",
		Payload:      `{"status":"completed"}`,
		TraceContext: traceContext,
	}

	b, _ := json.Marshal(completionEvent)
	return h.Redis.Publish(ctx, channel, b).Err()
}
