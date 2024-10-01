package cloudslog

import (
	"log/slog"
	"strings"
	"testing"

	examplev1 "go.einride.tech/protobuf-sensitive/gen/einride/sensitive/example/v1"
	"gotest.tools/v3/assert"
)

func TestHandler_redact(t *testing.T) {
	var b strings.Builder
	logger := slog.New(newHandler(&b, LoggerConfig{}))
	logger.Info("test", "example", &examplev1.ExampleMessage{
		DebugRedactedField: "foobar",
	})
	assert.Assert(t, strings.Contains(b.String(), `"debugRedactedField":"<redacted>"`), b.String())
}
