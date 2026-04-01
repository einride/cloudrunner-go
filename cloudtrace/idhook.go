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
//
// Deprecated: The cloudslog.Handler automatically injects trace fields from the OpenTelemetry
// span context into logging.googleapis.com/trace. Additionally, as per
// https://docs.cloud.google.com/trace/docs/trace-log-integration the preferred format for
// that field is now just the trace ID, making this separate hook redundant.
func IDHook(ctx context.Context, traceContext Context) context.Context {
	return cloudslog.With(ctx, slog.String(IDKey, traceContext.TraceID))
}
