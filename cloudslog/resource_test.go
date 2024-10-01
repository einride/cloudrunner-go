package cloudslog

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"gotest.tools/v3/assert"
)

func TestHandler_resource(t *testing.T) {
	var b strings.Builder
	logger := slog.New(newHandler(&b, LoggerConfig{}))
	r, err := resource.New(context.Background(), resource.WithAttributes(
		attribute.KeyValue{
			Key:   "foo",
			Value: attribute.StringValue("bar"),
		},
	))
	assert.NilError(t, err)
	logger.Info("test", "resource", r)
	assert.Assert(t, strings.Contains(b.String(), `"resource":{"foo":"bar"}`))
}
