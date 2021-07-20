package cloudzap

import (
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"gotest.tools/v3/assert"
)

func TestTrace(t *testing.T) {
	var buffer zaptest.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logger := zap.New(zapcore.NewCore(encoder, &buffer, zap.DebugLevel))
	logger.Info("test", Trace("foo", "bar"))
	assert.Assert(
		t,
		strings.Contains(
			buffer.Stripped(),
			`"logging.googleapis.com/trace":"projects/foo/traces/bar"`,
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
