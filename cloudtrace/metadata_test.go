package cloudtrace

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_Context(t *testing.T) {
	t.Parallel()
	t.Run("read from empty context", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		value, ok := GetContext(ctx)
		assert.Equal(t, false, ok)
		assert.DeepEqual(t, Context{}, value)
	})
	t.Run("set and read from context", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		ctx = SetContext(ctx, Context{
			TraceID: "traceId",
			SpanID:  "spanId",
			Sampled: true,
		})
		value, ok := GetContext(ctx)
		assert.Equal(t, true, ok)
		assert.DeepEqual(
			t,
			Context{
				TraceID: "traceId",
				SpanID:  "spanId",
				Sampled: true,
			},
			value,
		)
	})
}
