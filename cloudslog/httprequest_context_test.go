package cloudslog

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	ltype "google.golang.org/genproto/googleapis/logging/type"
	"gotest.tools/v3/assert"
)

func TestHandler_httpRequestFromContext(t *testing.T) {
	t.Run("included in error report context", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{ReportErrors: true}))
		ctx := WithHTTPRequest(context.Background(), &ltype.HttpRequest{
			RequestMethod: "GET",
			RequestUrl:    "/test/path",
			UserAgent:     "test-agent",
		})
		logger.ErrorContext(ctx, "something went wrong")
		got := b.String()
		assert.Assert(t, strings.Contains(got, `"requestMethod":"GET"`), got)
		assert.Assert(t, strings.Contains(got, `"requestUrl":"/test/path"`), got)
		assert.Assert(t, strings.Contains(got, `"userAgent":"test-agent"`), got)
	})

	t.Run("not included when no httpRequest on context", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{ReportErrors: true}))
		logger.ErrorContext(context.Background(), "something went wrong")
		got := b.String()
		assert.Assert(t, !strings.Contains(got, `"httpRequest"`), got)
	})

	t.Run("not included for non-error levels", func(t *testing.T) {
		var b strings.Builder
		logger := slog.New(newHandler(&b, LoggerConfig{ReportErrors: true}))
		ctx := WithHTTPRequest(context.Background(), &ltype.HttpRequest{
			RequestMethod: "GET",
			RequestUrl:    "/test/path",
		})
		logger.InfoContext(ctx, "just info")
		got := b.String()
		// httpRequest should not appear in the context group for non-error logs.
		assert.Assert(t, !strings.Contains(got, `"httpRequest"`), got)
	})
}
