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
	globalmetric "go.opentelemetry.io/otel/metric/global"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
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
	)
	globalmetric.SetMeterProvider(provider)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		if isUnsupportedSamplerErr(err) {
			// The OpenCensus bridge does not support all features from OpenCensus,
			// for example custom samplers which is used in some libraries.
			// The bridge presumably falls back to the configured sampler, so
			// this error can be ignored.
			//
			// See
			// https://pkg.go.dev/go.opentelemetry.io/otel/bridge/opencensus
			return
		}
		if logger, ok := cloudzap.GetLogger(ctx); ok {
			logger.Warn("metric exporter error", zap.Error(err))
		}
	}))
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
	if err != nil {
		return false
	}
	return strings.Contains(err.Error(), "unsupported sampler")
}
