package observability

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
)

// SlogWatermillAdapter implements watermill.LoggerAdapter using *slog.Logger.
type SlogWatermillAdapter struct {
	logger *slog.Logger
}

func NewSlogWatermillAdapter(logger *slog.Logger) *SlogWatermillAdapter {
	return &SlogWatermillAdapter{logger: logger}
}

func (a *SlogWatermillAdapter) Error(msg string, err error, fields watermill.LogFields) {
	a.log(context.Background(), slog.LevelError, msg, err, fields)
}

func (a *SlogWatermillAdapter) Info(msg string, fields watermill.LogFields) {
	a.log(context.Background(), slog.LevelInfo, msg, nil, fields)
}

func (a *SlogWatermillAdapter) Debug(msg string, fields watermill.LogFields) {
	a.log(context.Background(), slog.LevelDebug, msg, nil, fields)
}

func (a *SlogWatermillAdapter) Trace(msg string, fields watermill.LogFields) {
	// Slog doesn't have Trace level, mapping to Debug
	a.log(context.Background(), slog.LevelDebug, msg, nil, fields)
}

func (a *SlogWatermillAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	newLogger := a.logger.With(a.fieldsToArgs(fields)...)
	return &SlogWatermillAdapter{logger: newLogger}
}

func (a *SlogWatermillAdapter) log(ctx context.Context, level slog.Level, msg string, err error, fields watermill.LogFields) {
	args := a.fieldsToArgs(fields)
	if err != nil {
		args = append(args, "error", err)
	}
	a.logger.Log(ctx, level, msg, args...)
}

func (a *SlogWatermillAdapter) fieldsToArgs(fields watermill.LogFields) []any {
	var args []any
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
}
