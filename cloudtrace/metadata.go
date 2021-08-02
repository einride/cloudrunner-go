package cloudtrace

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// ContextHeader is the metadata key of the Cloud Trace context header.
const ContextHeader = "x-cloud-trace-context"

// FromIncomingContext returns the incoming Cloud Trace Context.
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
