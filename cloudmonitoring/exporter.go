package cloudmonitoring

import (
	"context"
	"fmt"
	"time"

	metricexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.einride.tech/cloudrunner/cloudzap"
	hostinstrumentation "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeinstrumentation "go.opentelemetry.io/contrib/instrumentation/runtime"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
)

// ExporterConfig configures the metrics exporter.
type ExporterConfig struct {
	Enabled                bool          `onGCE:"true"`
	Interval               time.Duration `default:"60s"`
	RuntimeInstrumentation bool          `onGCE:"true"`
	HostInstrumentation    bool          `onGCE:"true"`
}

// StartExporter starts the OpenTelemetry Cloud Monitoring exporter.
func StartExporter(
	ctx context.Context,
	exporterConfig ExporterConfig,
	resource *resource.Resource,
) (func(), error) {
	if !exporterConfig.Enabled {
		return func() {}, nil
	}
	projectID, ok := cloudruntime.ProjectID()
	if !ok {
		return nil, fmt.Errorf("start metric exporter: unknown project ID")
	}
	exporter, err := metricexporter.InstallNewPipeline(
		[]metricexporter.Option{
			metricexporter.WithProjectID(projectID),
			metricexporter.WithOnError(func(err error) {
				if logger, ok := cloudzap.GetLogger(ctx); ok {
					logger.Warn("metric exporter error", zap.Error(err))
				}
			}),
			metricexporter.WithInterval(exporterConfig.Interval),
		},
		controller.WithResource(resource),
	)
	if err != nil {
		return nil, fmt.Errorf("start metric exporter: %w", err)
	}
	shutdown := func() {
		if err := exporter.Stop(context.Background()); err != nil {
			if logger, ok := cloudzap.GetLogger(ctx); ok {
				logger.Error("error stopping metric exporter", zap.Error(err))
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
