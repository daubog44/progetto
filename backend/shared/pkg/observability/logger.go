package observability

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// ContextHandler is a custom slog handler that adds TraceID and SpanID to logs.
type ContextHandler struct {
	slog.Handler
}

// Handle adds trace_id and span_id to the record if they exist in the context.
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		spanContext := span.SpanContext()
		if spanContext.HasTraceID() {
			r.AddAttrs(slog.String("trace_id", spanContext.TraceID().String()))
		}
		if spanContext.HasSpanID() {
			r.AddAttrs(slog.String("span_id", spanContext.SpanID().String()))
		}
	}
	return h.Handler.Handle(ctx, r)
}

// LoggerWithContext wraps the default handler with the ContextHandler.
// Call this instead of the simple initLogger in init.go
func InitLogger() {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	handler := ContextHandler{Handler: baseHandler}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
