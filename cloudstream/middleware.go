package cloudstream

import (
	"context"

	"google.golang.org/grpc"
)

// ContextualServerStream wraps a "normal" grpc.Server stream but replaces the context.
// This is useful in for example middlewares.
type ContextualServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// NewContextualServerStream creates a new server stream, that uses the given context.
func NewContextualServerStream(ctx context.Context, root grpc.ServerStream) *ContextualServerStream {
	return &ContextualServerStream{
		ServerStream: root,
		ctx:          ctx,
	}
}

// Context returns the context for this server stream.
func (s *ContextualServerStream) Context() context.Context {
	return s.ctx
}
