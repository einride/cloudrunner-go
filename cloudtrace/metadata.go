package cloudtrace

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// ContextHeader is the metadata key of the Cloud Trace context header.
const ContextHeader = "x-cloud-trace-context"

// FromIncomingContext returns the incoming Cloud Trace Context.
// Deprecated: FromIncomingContext does not handle trace context coming from a HTTP server,
// use GetContext instead.
func FromIncomingContext(ctx context.Context) (Context, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return Context{}, false
	}
	values := md.Get(ContextHeader)
	if len(values) != 1 {
		return Context{}, false
	}
	var result Context
	if err := result.UnmarshalString(values[0]); err != nil {
		return Context{}, false
	}
	return result, true
}

type contextKey struct{}

// SetContext sets the cloud trace context to the provided context.
func SetContext(ctx context.Context, ctxx Context) context.Context {
	return context.WithValue(ctx, contextKey{}, ctxx)
}

// GetContext gets the cloud trace context from the provided context if it exists.
func GetContext(ctx context.Context) (Context, bool) {
	result, ok := ctx.Value(contextKey{}).(Context)
	if !ok {
		return Context{}, false
	}
	return result, true
}
