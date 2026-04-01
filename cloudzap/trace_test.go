package cloudzap

import (
	"strings"
	"testing"

	"go.uber.org/zap"         //nolint:gomodguard // cloudzap is a zap integration package
	"go.uber.org/zap/zapcore" //nolint:gomodguard // cloudzap is a zap integration package
	"go.uber.org/zap/zaptest" //nolint:gomodguard // cloudzap is a zap integration package
	"gotest.tools/v3/assert"
)

func TestTrace(t *testing.T) {
	var buffer zaptest.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logger := zap.New(zapcore.NewCore(encoder, &buffer, zap.DebugLevel))
	logger.Info("test", Trace("bar"))
	assert.Assert(
		t,
		strings.Contains(
			buffer.Stripped(),
			`"logging.googleapis.com/trace":"bar"`,
		),
	)
}

func TestSpanID(t *testing.T) {
	var buffer zaptest.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logger := zap.New(zapcore.NewCore(encoder, &buffer, zap.DebugLevel))
	logger.Info("test", SpanID("foo"))
	assert.Assert(
		t,
		strings.Contains(
			buffer.Stripped(),
			`"logging.googleapis.com/spanId":"foo"`,
		),
	)
}

func TestTraceSampled(t *testing.T) {
	var buffer zaptest.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logger := zap.New(zapcore.NewCore(encoder, &buffer, zap.DebugLevel))
	logger.Info("test", TraceSampled(true))
	assert.Assert(
		t,
		strings.Contains(
			buffer.Stripped(),
			`"logging.googleapis.com/trace_sampled":true`,
		),
	)
}
