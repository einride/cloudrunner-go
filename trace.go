package cloudrunner

import (
	"context"

	cloudtrace "go.einride.tech/cloudrunner/cloudtrace"
)

// IncomingTraceContext returns the Cloud Trace context from the incoming request metadata.
// Deprecated: Use opentelemetry trace.SpanContextFromContext instead.
func IncomingTraceContext(ctx context.Context) (cloudtrace.Context, bool) {
	return cloudtrace.FromIncomingContext(ctx)
}

// GetTraceContext returns the Cloud Trace context from the incoming request.
// Deprecated: Use opentelemetry trace.SpanContextFromContext instead.
func GetTraceContext(ctx context.Context) (cloudtrace.Context, bool) {
	return cloudtrace.GetContext(ctx)
}
