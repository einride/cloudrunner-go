package cloudotel

import (
	"context"
	"fmt"
	"strings"
	"time"

	metricexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.einride.tech/cloudrunner/cloudzap"
	hostinstrumentation "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeinstrumentation "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.14.0"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricExporterConfig configures the metrics exporter.
type MetricExporterConfig struct {
	Enabled                bool          `onGCE:"false"`
	Interval               time.Duration `default:"60s"`
	RuntimeInstrumentation bool          `onGCE:"true"`
	HostInstrumentation    bool          `onGCE:"true"`
}

// StartMetricExporter starts the OpenTelemetry Cloud Monitoring exporter.
func StartMetricExporter(
	ctx context.Context,
	exporterConfig MetricExporterConfig,
	resource *resource.Resource,
) (func(), error) {
	if !exporterConfig.Enabled {
		return func() {}, nil
	}
	projectID, ok := cloudruntime.ProjectID()
	if !ok {
		return nil, fmt.Errorf("start metric exporter: unknown project ID")
	}
	exporter, err := metricexporter.New(
		metricexporter.WithProjectID(projectID),
	)
	if err != nil {
		return nil, fmt.Errorf("new metric exporter: %w", err)
	}
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(exporterConfig.Interval))),
		sdkmetric.WithResource(resource),
		sdkmetric.WithView(
			// `net.sock.peer.port`, `net.port.peer` and `http.client_ip are high-cardinality attributes (essentially
			// one unique value per request) which causes failures when exporting metrics as the request limit
			// towards GCP is reached (200 time series per request).
			//
			// The following views masks these attributes from both otelhttp and otelgrpc so
			// that metrics can still be exported.
			// Based on https://github.com/open-telemetry/opentelemetry-go-contrib/issues/3071#issuecomment-1416137206
			maskInstrumentAttrs(
				"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
				semconv.NetPeerPortKey,
				semconv.NetSockPeerPortKey,
				semconv.HTTPClientIPKey,
			),
			maskInstrumentAttrs(
				"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
				semconv.NetPeerPortKey,
				semconv.NetSockPeerPortKey,
			),
		),
	)
	otel.SetMeterProvider(provider)
	shutdown := func() {
		if err := exporter.Shutdown(context.Background()); err != nil {
			if logger, ok := cloudzap.GetLogger(ctx); ok {
				const msg = "error stopping metric exporter, final metric export might have failed"
				switch status.Code(err) {
				case codes.InvalidArgument:
					// In case final export happens within the minimum frequency time from the previous export,
					// Cloud Monitoring API will fail with InvalidArgument. In that case, there's nothing to do
					// so only warn about it.
					logger.Warn(msg, zap.Error(err))
				default:
					logger.Error(msg, zap.Error(err))
				}
			}
		}
	}
	if exporterConfig.RuntimeInstrumentation {
		if err := runtimeinstrumentation.Start(); err != nil {
			shutdown()
			return nil, fmt.Errorf("start metric exporter: start runtime instrumentation: %w", err)
		}
	}
	if exporterConfig.HostInstrumentation {
		if err := hostinstrumentation.Start(); err != nil {
			shutdown()
			return nil, fmt.Errorf("start metric exporter: start host instrumentation: %w", err)
		}
	}
	return shutdown, nil
}

func isUnsupportedSamplerErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "unsupported sampler")
}

func maskInstrumentAttrs(instrumentScopeName string, attrs ...attribute.Key) sdkmetric.View {
	masked := make(map[attribute.Key]struct{})
	for _, attr := range attrs {
		masked[attr] = struct{}{}
	}
	return sdkmetric.NewView(
		sdkmetric.Instrument{Scope: instrumentation.Scope{Name: instrumentScopeName}},
		sdkmetric.Stream{
			AttributeFilter: func(value attribute.KeyValue) bool {
				_, ok := masked[value.Key]
				return !ok
			},
		},
	)
}
