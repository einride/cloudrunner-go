package cloudslog

import (
	"log/slog"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHandler(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))
		logger.Info("test")
		assert.Assert(t, strings.Contains(b.String(), "logging.googleapis.com/sourceLocation"))
	})
}
