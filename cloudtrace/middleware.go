package cloudtrace

import (
	"context"
	"net/http"

	"go.einride.tech/cloudrunner/cloudstream"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Middleware that ensures incoming traces are forwarded and included in logging.
type Middleware struct {
	// ProjectID of the project the service is running in.
	ProjectID string
	// TraceHook is an optional callback that gets called with the parsed trace context.
	TraceHook func(context.Context, Context) context.Context
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
	ctx = i.withInternalContext(ctx, values[0])
	ctx = i.withLogTracing(ctx, values[0])
	return handler(ctx, req)
}

// GRPCStreamServerInterceptor adds tracing metadata to streaming RPCs.
func (i *Middleware) GRPCStreamServerInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return handler(srv, ss)
	}
	values := md.Get(ContextHeader)
	if len(values) != 1 {
		return handler(srv, ss)
	}
	ctx := ss.Context()
	ctx = i.withOutgoingRequestTracing(ctx, values[0])
	ctx = i.withInternalContext(ctx, values[0])
	ctx = i.withLogTracing(ctx, values[0])
	return handler(srv, cloudstream.NewContextualServerStream(ctx, ss))
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
		ctx = i.withInternalContext(ctx, header)
		ctx = i.withLogTracing(ctx, header)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (i *Middleware) withOutgoingRequestTracing(ctx context.Context, header string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, ContextHeader, header)
}

func (i *Middleware) withInternalContext(ctx context.Context, header string) context.Context {
	var result Context
	if err := result.UnmarshalString(header); err != nil {
		return ctx
	}
	return SetContext(ctx, result)
}

func (i *Middleware) withLogTracing(ctx context.Context, header string) context.Context {
	var traceContext Context
	if err := traceContext.UnmarshalString(header); err != nil {
		return ctx
	}
	if i.TraceHook != nil {
		ctx = i.TraceHook(ctx, traceContext)
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
