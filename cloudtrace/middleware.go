package cloudtrace

import (
	"context"
	"net/http"

	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Middleware that ensures incoming traces are forwarded and included in logging.
type Middleware struct {
	// ProjectID of the project the service is running in.
	ProjectID string
}

// GRPCServerUnaryInterceptor provides unary RPC middleware for gRPC servers.
func (i *Middleware) GRPCServerUnaryInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}
	values := md.Get(ContextHeader)
	if len(values) != 1 {
		return handler(ctx, req)
	}
	ctx = i.withOutgoingRequestTracing(ctx, values[0])
	ctx = i.withLogTracing(ctx, values[0])
	return handler(ctx, req)
}

// HTTPServer provides middleware for HTTP servers.
func (i *Middleware) HTTPServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(ContextHeader)
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set(ContextHeader, header)
		ctx := i.withOutgoingRequestTracing(r.Context(), header)
		ctx = i.withLogTracing(ctx, header)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (i *Middleware) withOutgoingRequestTracing(ctx context.Context, header string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, ContextHeader, header)
}

func (i *Middleware) withLogTracing(ctx context.Context, header string) context.Context {
	var traceContext Context
	if err := traceContext.UnmarshalString(header); err != nil {
		return ctx
	}
	fields := make([]zap.Field, 0, 3)
	fields = append(fields, cloudzap.Trace(i.ProjectID, traceContext.TraceID))
	if traceContext.SpanID != "" {
		fields = append(fields, cloudzap.SpanID(traceContext.SpanID))
	}
	if traceContext.Sampled {
		fields = append(fields, cloudzap.TraceSampled(traceContext.Sampled))
	}
	return cloudzap.WithLoggerFields(ctx, fields...)
}
