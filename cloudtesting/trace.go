package cloudtesting

import (
	"context"

	"go.einride.tech/cloudrunner/cloudtrace"
	"google.golang.org/grpc/metadata"
)

// WithIncomingTraceContext returns a new context with the specified trace.
func WithIncomingTraceContext(ctx context.Context, traceContext cloudtrace.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	return metadata.NewIncomingContext(
		ctx,
		metadata.Join(md, metadata.Pairs(cloudtrace.ContextHeader, traceContext.String())),
	)
}
