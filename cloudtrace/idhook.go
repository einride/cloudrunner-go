package cloudtrace

import (
	"context"
	"log/slog"

	"go.einride.tech/cloudrunner/cloudslog"
)

// IDKey is the log entry key for trace IDs.
// Experimental: May be removed in a future update.
const IDKey = "traceId"

// IDHook adds the trace ID (without the full trace resource name) to the request logger.
// The trace ID can be used to filter on logs for the same trace across multiple projects.
// Experimental: May be removed in a future update.
func IDHook(ctx context.Context, traceContext Context) context.Context {
	return cloudslog.With(ctx, slog.String(IDKey, traceContext.TraceID))
}
