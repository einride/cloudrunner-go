package cloudotel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	metricexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.einride.tech/cloudrunner/cloudruntime"
	hostinstrumentation "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeinstrumentation "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	ocbridge "go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// MetricExporterConfig configures the metrics exporter.
type MetricExporterConfig struct {
	Enabled                bool          `onGCE:"false"`
	Interval               time.Duration `default:"60s"`
	RuntimeInstrumentation bool          `onGCE:"true"`
	HostInstrumentation    bool          `onGCE:"true"`
	OpenCensusProducer     bool          `default:"false"`
	// DropMetrics is a list of metric names to drop. Supports wildcards (e.g., "http.client.*").
	// This can be used to reduce cost and cardinality by excluding unwanted metrics.
	DropMetrics []string
}

// StartMetricExporter starts the OpenTelemetry Cloud Monitoring exporter.
func StartMetricExporter(
	ctx context.Context,
	exporterConfig MetricExporterConfig,
	resource *resource.Resource,
) (func(context.Context) error, error) {
	if !exporterConfig.Enabled {
		return func(context.Context) error { return nil }, nil
	}
	projectID, ok := cloudruntime.ResolveProjectID(ctx)
	if !ok {
		return nil, fmt.Errorf("start metric exporter: unknown project ID")
	}
	exporter, err := metricexporter.New(
		metricexporter.WithProjectID(projectID),
	)
	if err != nil {
		return nil, fmt.Errorf("new metric exporter: %w", err)
	}
	readerOpts := []sdkmetric.PeriodicReaderOption{
		sdkmetric.WithInterval(exporterConfig.Interval),
	}
	if exporterConfig.OpenCensusProducer {
		readerOpts = append(readerOpts, sdkmetric.WithProducer(ocbridge.NewMetricProducer()))
	}
	reader := sdkmetric.NewPeriodicReader(exporter, readerOpts...)
	// Build views for the meter provider.
	// `net.sock.peer.port`, `net.port.peer` and `http.client_ip are high-cardinality attributes (essentially
	// one unique value per request) which causes failures when exporting metrics as the request limit
	// towards GCP is reached (200 time series per request).
	//
	// The following views masks these attributes from both otelhttp and otelgrpc so
	// that metrics can still be exported.
	// Based on https://github.com/open-telemetry/opentelemetry-go-contrib/issues/3071#issuecomment-1416137206
	views := []sdkmetric.View{
		maskInstrumentAttrs(
			"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
			semconv.NetPeerPortKey,
			semconv.NetSockPeerPortKey,
			attribute.Key("http.client_ip"),
		),
		maskInstrumentAttrs(
			"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
			semconv.NetPeerPortKey,
			semconv.NetSockPeerPortKey,
		),
	}
	// Add drop views for metrics specified in DropMetrics.
	for _, name := range exporterConfig.DropMetrics {
		views = append(views, dropMetricView(name))
	}
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(resource),
		sdkmetric.WithView(views...),
	)
	otel.SetMeterProvider(provider)
	shutdown := func(ctx context.Context) error {
		if err := provider.Shutdown(ctx); err != nil {
			return fmt.Errorf("error stopping metric provider, final metric export might have failed: %v", err)
		}
		return nil
	}
	if exporterConfig.RuntimeInstrumentation {
		if err := runtimeinstrumentation.Start(); err != nil {
			return nil, errors.Join(
				shutdown(ctx),
				fmt.Errorf("start metric exporter: start runtime instrumentation: %w", err),
			)
		}
	}
	if exporterConfig.HostInstrumentation {
		if err := hostinstrumentation.Start(); err != nil {
			return nil, errors.Join(
				shutdown(ctx),
				fmt.Errorf("start metric exporter: start host instrumentation: %w", err),
			)
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

func dropMetricView(name string) sdkmetric.View {
	return sdkmetric.NewView(
		sdkmetric.Instrument{Name: name},
		sdkmetric.Stream{Aggregation: sdkmetric.AggregationDrop{}},
	)
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
