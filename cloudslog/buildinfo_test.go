package cloudslog

import (
	"log/slog"
	"runtime/debug"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHandler_buildInfoValue(t *testing.T) {
	var b strings.Builder
	logger := slog.New(newHandler(&b, LoggerConfig{}))
	buildInfo, ok := debug.ReadBuildInfo()
	assert.Assert(t, ok)
	logger.Info("test", "buildInfo", buildInfo)
	t.Log(b.String())
	assert.Assert(t, strings.Contains(b.String(), `"buildInfo":{"mainPath":`))
}
