package cloudtesting

import (
	"context"
	"testing"

	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap/zaptest"
)

// WithLogger returns a new context with a test logger.
func WithLogger(ctx context.Context, t *testing.T, opts ...zaptest.LoggerOption) context.Context {
	return cloudzap.WithLogger(ctx, zaptest.NewLogger(t, opts...))
}
