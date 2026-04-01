package cloudzap

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"         //nolint:gomodguard // cloudzap is a zap integration package
	"go.uber.org/zap/zapcore" //nolint:gomodguard // cloudzap is a zap integration package
	"go.uber.org/zap/zaptest" //nolint:gomodguard // cloudzap is a zap integration package
	"gotest.tools/v3/assert"
)

func TestResource(t *testing.T) {
	var buffer zaptest.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	input, err := resource.New(context.Background(), resource.WithAttributes(
		attribute.KeyValue{
			Key:   "foo",
			Value: attribute.StringValue("bar"),
		},
	))
	assert.NilError(t, err)
	logger := zap.New(zapcore.NewCore(encoder, &buffer, zap.DebugLevel))
	logger.Info("test", Resource("resource", input))
	assert.Assert(
		t,
		strings.Contains(
			buffer.Stripped(),
			`"resource":{"foo":"bar"}`,
		),
	)
}
