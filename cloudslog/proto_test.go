package cloudslog

import (
	"log/slog"
	"strings"
	"testing"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"gotest.tools/v3/assert"
)

func TestHandler_proto(t *testing.T) {
	t.Run("message", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		logger.Info("test", "httpRequest", &ltype.HttpRequest{
			RequestMethod: "GET",
		})
		assert.Assert(t, strings.Contains(b.String(), "requestMethod"))
	})

	t.Run("enum", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		logger.Info("test", "logEntry", &loggingpb.LogEntry{
			Severity: ltype.LogSeverity_INFO,
		})
		assert.Assert(t, strings.Contains(b.String(), `"severity":"INFO"`))
	})
}
