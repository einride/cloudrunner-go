package cloudslog

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHandler_withContextAttributes(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		ctx := context.Background()
		ctx = With(ctx, "foo", "bar")
		logger.InfoContext(ctx, "test")
		assert.Assert(t, strings.Contains(b.String(), `"foo":"bar"`))
	})

	t.Run("attrs", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		ctx := context.Background()
		ctx = With(ctx, slog.String("foo", "bar"), slog.Int("baz", 3))
		logger.InfoContext(ctx, "test")
		assert.Assert(t, strings.Contains(b.String(), `"foo":"bar","baz":3`))
	})

	t.Run("multiple", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		ctx := context.Background()
		ctx = With(ctx, "foo", "bar")
		ctx = With(ctx, "lorem", "ipsum")
		logger.InfoContext(ctx, "test")
		assert.Assert(t, strings.Contains(b.String(), `"foo":"bar","lorem":"ipsum"`))
	})

	t.Run("scoped", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		ctx := context.Background()
		ctx = With(ctx, "foo", "bar")
		_ = With(ctx, "lorem", "ipsum")
		logger.InfoContext(ctx, "test")
		assert.Assert(t, strings.Contains(b.String(), `"foo":"bar"`))
		assert.Assert(t, !strings.Contains(b.String(), `"lorem":"ipsum"`))
	})
}
