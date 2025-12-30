package sse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/username/progetto/shared/pkg/jwtutil"
	"github.com/username/progetto/shared/pkg/presence"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Event struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type Handler struct {
	clients    map[string]chan Event
	lock       sync.RWMutex
	rdb        *redis.Client
	instanceID string
	jwtSecret  []byte
}

func NewHandler(rdb *redis.Client, jwtSecret string) *Handler {
	id := os.Getenv("HOSTNAME")
	if id == "" {
		id = uuid.New().String()
	}

	return &Handler{
		clients:    make(map[string]chan Event),
		rdb:        rdb,
		instanceID: id,
		jwtSecret:  []byte(jwtSecret),
	}
}

// RegisterRoutes registers the SSE endpoint
func (h *Handler) RegisterRoutes(api huma.API) {
	// Huma doesn't support streaming response easily with OpenAPI typed output yet for SSE easily without custom handler
	// So we use standard http.HandlerFunc wrapper or register raw route in Chi router if possible.
	// Huma can verify input though.

	// We'll use the underlying router access if needed, or Huma's Register with low-level writer access
	// But Huma handlers return structs.
	// Strategy: Register a raw handler on the underlying router (Chi) in main.go,
	// because SSE is a long-running connection, not a simple Request-Response.
}

// ServeHTTP implements the SSE endpoint
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 2. Auth via JWT
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}

	userID, err := jwtutil.ValidateToken(tokenString, h.jwtSecret)
	if err != nil {
		if errors.Is(err, jwtutil.ErrExpiredToken) {
			slog.Warn("token expired", "error", err)
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}
		slog.Error("failed to validate token", "error", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	clientChan := make(chan Event, 10)
	h.addClient(userID, clientChan)

	// Register online presence
	if err := h.setPresence(r.Context(), userID, "online", nil); err != nil {
		slog.Error("failed to register presence", "user_id", userID, "error", err)
	}

	defer func() {
		h.removeClient(userID)
		// Mark as offline on disconnect
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		h.setOffline(ctx, userID)
	}()

	// 3. Keep connection open
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial ping
	fmt.Fprintf(w, "event: connected\ndata: \"connected\"\n\n")
	flusher.Flush()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case event := <-clientChan:
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, event.Payload)
			flusher.Flush()
		case <-time.After(15 * time.Second):
			// Keep-alive heartbeat
			fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()

			// Refresh presence heartbeat
			go h.updateLastSeen(context.Background(), userID)
		}
	}
}

func (h *Handler) updateLastSeen(ctx context.Context, userID string) {
	h.setPresence(ctx, userID, "online", nil)
}

func (h *Handler) setOffline(ctx context.Context, userID string) {
	now := time.Now()
	h.setPresence(ctx, userID, "offline", &now)
}

func (h *Handler) setPresence(ctx context.Context, userID string, status string, disconnectedAt *time.Time) error {
	key := fmt.Sprintf("user_presence:%s", userID)

	p := presence.UserPresence{
		InstanceID:     h.instanceID,
		Status:         status,
		UpdatedAt:      time.Now(),
		DisconnectedAt: disconnectedAt,
	}

	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return h.rdb.Set(ctx, key, b, 7*24*time.Hour).Err()
}

// SubscribeToRedis starts listening for events from Redis for this specific gateway instance
func (h *Handler) SubscribeToRedis(ctx context.Context) error {
	channel := fmt.Sprintf("gateway_events:%s", h.instanceID)
	pubsub := h.rdb.Subscribe(ctx, channel)
	defer pubsub.Close()

	slog.Info("Subscribed to targeted Redis channel", "channel", channel)

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-ch:
			var sseEvent presence.TargetedEvent
			if err := json.Unmarshal([]byte(msg.Payload), &sseEvent); err != nil {
				slog.Error("failed to unmarshal redis event", "error", err)
				continue
			}

			traceCtx := ctx
			if sseEvent.TraceContext != nil {
				// Extract Trace Context
				carrier := propagation.MapCarrier(sseEvent.TraceContext)
				traceCtx = otel.GetTextMapPropagator().Extract(ctx, carrier)
			}

			// Start Span to link traces
			tracer := otel.Tracer("gateway-service")
			spanCtx, span := tracer.Start(traceCtx, "process_redis_event", trace.WithSpanKind(trace.SpanKindConsumer))
			defer span.End()

			// Targeted delivery: check if the user is actually on this instance
			h.Broadcast(spanCtx, sseEvent.UserID, sseEvent.Type, sseEvent.Payload)
		}
	}
}

func (h *Handler) Broadcast(ctx context.Context, userID string, eventType, payload string) {
	h.lock.RLock()
	ch, ok := h.clients[userID]
	h.lock.RUnlock()

	if ok {
		select {
		case ch <- Event{Type: eventType, Payload: payload}:
			slog.InfoContext(ctx, "Broadcasted event to user", "user_id", userID, "type", eventType)
		default:
			slog.WarnContext(ctx, "Client channel full, dropping event", "user_id", userID)
		}
	} else {
		// Debug log might not be needed always, but if enabled useful to trace
		slog.DebugContext(ctx, "User not connected, skipping broadcast", "user_id", userID)
	}
}

func (h *Handler) addClient(userID string, ch chan Event) {
	h.lock.Lock()
	defer h.lock.Unlock()
	// Close old connection if exists? Or allow multiple?
	// For now simple overwrite (last wins) or we need a slice of channels for multiple tabs.
	// Simplifying to single connection per user.
	if old, ok := h.clients[userID]; ok {
		close(old)
	}
	h.clients[userID] = ch
}

func (h *Handler) removeClient(userID string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.clients, userID)
}
