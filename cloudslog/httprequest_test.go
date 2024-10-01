package cloudslog

import (
	"log/slog"
	"strings"
	"testing"

	ltype "google.golang.org/genproto/googleapis/logging/type"
	"gotest.tools/v3/assert"
)

func TestHandler_httpRequest(t *testing.T) {
	var b strings.Builder
	logger := slog.New(newHandler(&b, LoggerConfig{}))
	logger.Info("test", "httpRequest", &ltype.HttpRequest{
		RequestMethod: "GET",
		RequestUrl:    "/foo/bar",
	})
	assert.Assert(t, strings.Contains(b.String(), `"requestUrl":"/foo/bar"`))
}
