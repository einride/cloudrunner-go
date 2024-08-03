package cloudslog

import (
	"context"
	"log/slog"
	"sync"
)

type contextKey struct{}

type contextValue struct {
	mu    sync.Mutex
	attrs []slog.Attr
}

// WithLoggerFields attaches structured fields to a new logger in the returned child context.
func WithAttributes(ctx context.Context, attrs ...slog.Attr) context.Context {
	value, ok := ctx.Value(contextKey{}).(*contextValue)
	if !ok {
		value = &contextValue{}
		ctx = context.WithValue(ctx, contextKey{}, value)
	}
	value.mu.Lock()
	defer value.mu.Unlock()
	value.attrs = append(value.attrs, attrs...)
	return ctx
}
