package sse

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type Event struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type Handler struct {
	clients map[string]chan Event
	lock    sync.RWMutex
}

func NewHandler() *Handler {
	return &Handler{
		clients: make(map[string]chan Event),
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

	// 2. Auth (extract UserID from query or context if middleware set it)
	// For simplicity, let's assume UserID is passed as query param `user_id` used for targeting
	// In production -> extract from JWT in context
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	clientChan := make(chan Event, 10)
	h.addClient(userID, clientChan)
	defer h.removeClient(userID)

	// 3. Keep connection open
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial ping/connection established
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
