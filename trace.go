package cloudrunner

import (
	"context"

	"go.einride.tech/cloudrunner/cloudtrace"
)

// IncomingTraceContext returns the Cloud Trace context from the incoming request metadata.
func IncomingTraceContext(ctx context.Context) (cloudtrace.Context, bool) {
	return cloudtrace.FromIncomingContext(ctx)
}
