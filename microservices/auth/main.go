package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

// ClerkWebhookInput represents the payload from Clerk
type ClerkWebhookInput struct {
	Body struct {
		Data jsonRawMessage `json:"data"`
		Type string         `json:"type" doc:"Clerk event type (e.g. user.created)"`
	}
}

// Simple wrapper for raw message to preserve data for logging/processing
type jsonRawMessage []byte

func (m *jsonRawMessage) UnmarshalJSON(data []byte) error {
	*m = append((*m)[0:0], data...)
	return nil
}

func (m jsonRawMessage) MarshalJSON() ([]byte, error) {
	return m, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Auth Service", "1.0.0"))

	// Health check
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
	}, func(ctx context.Context, input *struct{}) (*struct{ Body string }, error) {
		return &struct{ Body string }{Body: "OK"}, nil
	})

	// Clerk Webhook Handler
	huma.Register(api, huma.Operation{
		OperationID: "clerk-webhook",
		Method:      http.MethodPost,
		Path:        "/webhooks/clerk",
		Summary:     "Clerk Webhook Handler",
		Description: "Receive and process events from Clerk (e.g. user registration, login)",
		Tags:        []string{"Webhooks"},
	}, func(ctx context.Context, input *ClerkWebhookInput) (*struct{ Status int }, error) {
		logger.Info("Received Clerk webhook",
			"type", input.Body.Type,
			"data", string(input.Body.Data),
		)

		// TODO: Implement actual synchronization logic (e.g. save user to DB, publish event to Kafka)

		return &struct{ Status int }{Status: http.StatusOK}, nil
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Auth service listening", "port", port)
	http.ListenAndServe(":"+port, router)
}
