package cloudmonitoring

import (
	"context"

	"go.einride.tech/cloudrunner/cloudotel"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Deprecated: use cloudotel.ExporterConfig.
type ExporterConfig = cloudotel.MetricExporterConfig

// StartExporter starts the OpenTelemetry Cloud Monitoring exporter.
// Deprecated: use cloudotel.StartMetricExporter.
func StartExporter(
	ctx context.Context,
	exporterConfig ExporterConfig,
	resource *resource.Resource,
) (func(), error) {
	return cloudotel.StartMetricExporter(ctx, exporterConfig, resource)
}
