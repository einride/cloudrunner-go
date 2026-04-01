package cloudzap

import (
	"context"

	"go.uber.org/zap" //nolint:gomodguard // cloudzap is a zap integration package
)

type loggerContextKey struct{}

// WithLogger adds a logger to the current context.
//
// Deprecated: cloudrunner.Run configures the default slog logger with cloudslog.Handler.
// Use slog.InfoContext, slog.WarnContext, etc. instead of a context-based zap logger.
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// GetLogger returns the logger for the current context.
//
// Deprecated: Use slog.InfoContext, slog.WarnContext, etc. with the default slog logger instead.
func GetLogger(ctx context.Context) (*zap.Logger, bool) {
	logger, ok := ctx.Value(loggerContextKey{}).(*zap.Logger)
	return logger, ok
}

// WithLoggerFields attaches structured fields to a new logger in the returned child context.
//
// Deprecated: Use cloudslog.With to attach attributes to the context instead.
func WithLoggerFields(ctx context.Context, fields ...zap.Field) context.Context {
	logger, ok := ctx.Value(loggerContextKey{}).(*zap.Logger)
	if !ok {
		return ctx
	}
	return WithLogger(ctx, logger.With(fields...))
}
