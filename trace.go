package cloudrunner

import (
	"context"

	"go.einride.tech/cloudrunner/cloudtrace"
)

// IncomingTraceContext returns the Cloud Trace context from the incoming request metadata.
// Deprecated: Use GetTraceContext instead.
func IncomingTraceContext(ctx context.Context) (cloudtrace.Context, bool) {
	return cloudtrace.FromIncomingContext(ctx)
}

// GetTraceContext returns the Cloud Trace context from the incoming request.
func GetTraceContext(ctx context.Context) (cloudtrace.Context, bool) {
	return cloudtrace.GetContext(ctx)
}
