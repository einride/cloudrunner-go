package cloudzap

import (
	"context"

	"go.uber.org/zap"
)

type loggerContextKey struct{}

// WithLogger adds a logger to the current context.
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// GetLogger returns the logger for the current context.
func GetLogger(ctx context.Context) (*zap.Logger, bool) {
	logger, ok := ctx.Value(loggerContextKey{}).(*zap.Logger)
	return logger, ok
}

// WithLoggerFields attaches structured fields to a new logger in the returned child context.
func WithLoggerFields(ctx context.Context, fields ...zap.Field) context.Context {
	logger, ok := ctx.Value(loggerContextKey{}).(*zap.Logger)
	if !ok {
		return ctx
	}
	return WithLogger(ctx, logger.With(fields...))
}
