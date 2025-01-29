package cloudotel

import (
	"context"
	"net/http"

	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"go.einride.tech/cloudrunner/cloudstream"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TraceHook func(context.Context, trace.SpanContext) context.Context

// TraceMiddleware that ensures incoming traces are forwarded and included in logging.
type TraceMiddleware struct {
	// ProjectID of the project the service is running in.
	ProjectID string
	// TraceHook is an optional callback that gets called with the parsed trace context.
	TraceHook TraceHook
	// propagator is a opentelemetry trace propagator
	propagator propagation.TextMapPropagator
}

func NewTraceMiddleware() TraceMiddleware {
	propagator := propagation.NewCompositeTextMapPropagator(
		gcppropagator.CloudTraceFormatPropagator{},
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return TraceMiddleware{
		TraceHook:  TraceIDHook,
		propagator: propagator,
	}
}

// GRPCServerUnaryInterceptor provides unary RPC middleware for gRPC servers.
func (i *TraceMiddleware) GRPCServerUnaryInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}
	carrier := propagation.HeaderCarrier(md)
	ctx = i.propagator.Extract(ctx, carrier)
	ctx = i.withLogTracing(ctx, trace.SpanContextFromContext(ctx))
	return handler(ctx, req)
}

// GRPCStreamServerInterceptor adds tracing metadata to streaming RPCs.
func (i *TraceMiddleware) GRPCStreamServerInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return handler(srv, ss)
	}
	ctx := ss.Context()
	carrier := propagation.HeaderCarrier(md)
	ctx = i.propagator.Extract(ctx, carrier)
	ctx = i.withLogTracing(ctx, trace.SpanContextFromContext(ctx))
	return handler(srv, cloudstream.NewContextualServerStream(ctx, ss))
}

// HTTPServer provides middleware for HTTP servers.
func (i *TraceMiddleware) HTTPServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		carrier := propagation.HeaderCarrier(r.Header)
		ctx := i.propagator.Extract(r.Context(), carrier)
		ctx = i.withLogTracing(ctx, trace.SpanContextFromContext(ctx))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (i *TraceMiddleware) withLogTracing(ctx context.Context, spanCtx trace.SpanContext) context.Context {
	if i.TraceHook != nil {
		ctx = i.TraceHook(ctx, spanCtx)
	}
	fields := make([]zap.Field, 0, 3)
	fields = append(fields, cloudzap.Trace(i.ProjectID, spanCtx.TraceID().String()))
	if spanCtx.SpanID().String() != "" {
		fields = append(fields, cloudzap.SpanID(spanCtx.SpanID().String()))
	}
	if spanCtx.IsSampled() {
		fields = append(fields, cloudzap.TraceSampled(spanCtx.IsSampled()))
	}
	return cloudzap.WithLoggerFields(ctx, fields...)
}
