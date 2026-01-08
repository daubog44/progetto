package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"github.com/username/progetto/notification-service/internal/presencestore"
	"github.com/username/progetto/shared/pkg/presence"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type UserHandler struct {
	Redis    *redis.Client
	Presence *presencestore.Store
	Logger   *slog.Logger
}

func NewUserHandler(rdb *redis.Client, store *presencestore.Store) *UserHandler {
	return &UserHandler{
		Redis:    rdb,
		Presence: store,
		Logger:   slog.Default().With("component", "user_handler"),
	}
}

func (h *UserHandler) HandleUserCreated(msg *message.Message) error {
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil
	}

	key := fmt.Sprintf("registration:%s:pending_syncs", payload.UserID)
	pipe := h.Redis.Pipeline()
	pipe.SAdd(msg.Context(), key, "post", "social", "search")
	pipe.Expire(msg.Context(), key, 1*time.Minute)

	if _, err := pipe.Exec(msg.Context()); err != nil {
		return err
	}

	h.Logger.Info("initialized tracking for user registration", "user_id", payload.UserID)
	return nil
}

// HandleFailure constructs a compensation message for user creation failure.
func (h *UserHandler) HandleFailure(err error, msg *message.Message) (string, *message.Message, error) {
	// Try to get UserID from Metadata first (if passed), otherwise from Payload
	userID := msg.Metadata.Get("user_id")
	if userID == "" {
		var payload struct {
			UserID string `json:"user_id"`
		}
		// Best effort unmarshal
		_ = json.Unmarshal(msg.Payload, &payload)
		userID = payload.UserID
	}

	h.Logger.ErrorContext(msg.Context(), "handling failure for user_created", "user_id", userID, "error", err)

	failurePayload := struct {
		UserID string `json:"user_id"`
		Reason string `json:"reason"`
		Source string `json:"source"`
	}{
		UserID: userID,
		Reason: err.Error(),
		Source: "notification-service",
	}

	bytes, marshalErr := json.Marshal(failurePayload)
	if marshalErr != nil {
		return "", nil, marshalErr
	}

	failMsg := message.NewMessage(watermill.NewUUID(), bytes)
	failMsg.Metadata.Set("user_id", userID)

	// Return the topic 'user_creation_failed' and the message
	return "user_creation_failed", failMsg, nil
}

func (h *UserHandler) HandleUserSynced(msg *message.Message) error {
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
	case "user_synced_search":
		service = "search"
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

func (h *UserHandler) triggerOnboardingCompleted(ctx context.Context, userID string) error {
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
