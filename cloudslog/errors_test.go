package cloudslog

import (
	"errors"
	"log/slog"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestErrors(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{}))

		logger.Info("test", Errors([]error{errors.New("test_error")}))
		assert.Assert(t, strings.Contains(b.String(), "test_error"))
	})
}
