package cloudslog

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"gotest.tools/v3/assert"
)

func TestHandler(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		logger.Info("test")
		assert.Assert(t, strings.Contains(b.String(), "logging.googleapis.com/sourceLocation"))
	})

	t.Run("proto", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		logger.Info("test", "httpRequest", &ltype.HttpRequest{
			RequestMethod: "GET",
		})
		assert.Assert(t, strings.Contains(b.String(), "requestMethod"))
	})

	t.Run("resource", func(t *testing.T) {
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
	})

	t.Run("buildInfo", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		buildInfo, _ := debug.ReadBuildInfo()
		logger.Info("test", "buildInfo", buildInfo)
		assert.Assert(t, strings.Contains(b.String(), `"goVersion"`))
	})
}
