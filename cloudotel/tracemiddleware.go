package cloudotel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"go.einride.tech/cloudrunner/cloudpubsub"
	"go.einride.tech/cloudrunner/cloudstream"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	// EnablePubsubTracing, disabled by default, reads trace parent from Pub/Sub message attributes.
	EnablePubsubTracing bool
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
		if i.EnablePubsubTracing {
			// Check if it is a Pub/Sub message and propagate tracing if exists.
			ctx = propagatePubsubTracing(ctx, r)
		}
		ctx = i.withLogTracing(ctx, trace.SpanContextFromContext(ctx))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (i *TraceMiddleware) withLogTracing(ctx context.Context, spanCtx trace.SpanContext) context.Context {
	if i.TraceHook != nil {
		ctx = i.TraceHook(ctx, spanCtx)
	}
	// Trace fields are automatically added by the slog handler from the span context
	return ctx
}

func propagatePubsubTracing(ctx context.Context, r *http.Request) context.Context {
	if r.Method != http.MethodPost {
		return ctx
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ctx
	}
	// Replace the original request body, so it can be read again.
	r.Body = io.NopCloser(bytes.NewReader(body))
	pubsubPayload, err := tryUnmarshalAsPubsubPayload(body)
	if err != nil {
		return ctx
	}
	pubsubMessage := pubsubPayload.BuildPubSubMessage()
	ctx = injectTracingFromPubsubMsg(ctx, &pubsubMessage)
	return ctx
}

func tryUnmarshalAsPubsubPayload(body []byte) (cloudpubsub.Payload, error) {
	var payload cloudpubsub.Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		return payload, err
	}
	if !payload.IsValid() {
		return payload, errors.New("not a pubsub payload")
	}
	return payload, nil
}

func injectTracingFromPubsubMsg(ctx context.Context, pubsubMessage *pubsubpb.PubsubMessage) context.Context {
	tc := propagation.TraceContext{}
	ctx = tc.Extract(ctx, pubsub.NewMessageCarrierFromPB(pubsubMessage))
	carrier := make(propagation.MapCarrier)
	tc.Inject(ctx, &carrier)
	return ctx
}
