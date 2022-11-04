package cloudmonitoring

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Metric names are based on OTEL semantic conventions for metrics.
// See:
// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics
const (
	serverRequestDurationMetricName = "rpc.cloudrunner.server.duration"
	clientRequestDurationMetricName = "rpc.cloudrunner.client.duration"

	// there is no rpc_count equivalent int OTEL semantic conventions yet.
	serverRequestCountMetricName = "rpc.server.rpc_count"
	clientRequestCountMetricName = "rpc.client.rpc_count"
)

func NewMetricMiddleware() (MetricMiddleware, error) {
	meter := global.MeterProvider().Meter("cloudrunner-go/cloudmonitoring")

	serverRequestCount, err := meter.SyncInt64().Counter(
		serverRequestCountMetricName,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Count of RPCs received by a gRPC server."),
	)
	if err != nil {
		return MetricMiddleware{}, fmt.Errorf("create server request count counter: %w", err)
	}
	serverRequestDuration, err := meter.SyncInt64().Histogram(
		serverRequestDurationMetricName,
		instrument.WithUnit(unit.Milliseconds),
		instrument.WithDescription("Duration of RPCs received by a gRPC server."),
	)
	if err != nil {
		return MetricMiddleware{}, fmt.Errorf("create server request duration histogram: %w", err)
	}
	clientRequestCount, err := meter.SyncInt64().Counter(
		clientRequestCountMetricName,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Count of RPCs sent by a gRPC client."),
	)
	if err != nil {
		return MetricMiddleware{}, fmt.Errorf("create client request count counter: %w", err)
	}
	clientRequestDuration, err := meter.SyncInt64().Histogram(
		clientRequestDurationMetricName,
		instrument.WithUnit(unit.Milliseconds),
		instrument.WithDescription("Duration of RPCs sent by a gRPC client."),
	)
	if err != nil {
		return MetricMiddleware{}, fmt.Errorf("create client request duration histogram: %w", err)
	}
	return MetricMiddleware{
		serverRequestCount:    serverRequestCount,
		serverRequestDuration: serverRequestDuration,
		clientRequestCount:    clientRequestCount,
		clientRequestDuration: clientRequestDuration,
	}, nil
}

type MetricMiddleware struct {
	serverRequestCount    syncint64.Counter
	serverRequestDuration syncint64.Histogram
	clientRequestCount    syncint64.Counter
	clientRequestDuration syncint64.Histogram
}

// GRPCUnaryServerInterceptor implements grpc.UnaryServerInterceptor and
// emits metrics for request count and request duration when a gRPC server
// receives requests.
func (m *MetricMiddleware) GRPCUnaryServerInterceptor(
	ctx context.Context,
	request interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	startTime := time.Now()
	response, err := handler(ctx, request)
	duration := time.Since(startTime)
	code := status.Code(err)

	attrs := rpcAttrs(info.FullMethod, code)
	m.serverRequestCount.Add(ctx, 1, attrs...)
	m.serverRequestDuration.Record(ctx, duration.Milliseconds(), attrs...)
	return response, err
}

// GRPCUnaryClientInterceptor provides request logging as a grpc.UnaryClientInterceptor.
func (m *MetricMiddleware) GRPCUnaryClientInterceptor(
	ctx context.Context,
	fullMethod string,
	request interface{},
	response interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	startTime := time.Now()
	err := invoker(ctx, fullMethod, request, response, cc, opts...)
	code := status.Code(err)
	duration := time.Since(startTime)

	attrs := rpcAttrs(fullMethod, code)
	m.clientRequestCount.Add(ctx, 1, attrs...)
	m.clientRequestDuration.Record(ctx, duration.Milliseconds(), attrs...)
	return err
}

func rpcAttrs(fullMethod string, code codes.Code) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 5)
	attrs = append(
		attrs,
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCGRPCStatusCodeKey.Int64(int64(code)),
		// Google Cloud Monitoring does not recognize semconv status code enum,
		// so add an attributes with string representation of status code.
		attribute.Stringer("rpc.grpc.code", code),
	)
	if service, method, ok := splitFullMethod(fullMethod); ok {
		attrs = append(
			attrs,
			semconv.RPCServiceKey.String(service),
			semconv.RPCMethodKey.String(method),
		)
	}
	return attrs
}

func splitFullMethod(fullMethod string) (service, method string, ok bool) {
	serviceAndMethod := strings.SplitN(strings.TrimPrefix(fullMethod, "/"), "/", 2)
	if len(serviceAndMethod) != 2 {
		return "", "", false
	}
	return serviceAndMethod[0], serviceAndMethod[1], true
}
