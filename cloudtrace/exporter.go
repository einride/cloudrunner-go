package cloudtrace

import (
	"context"

	"go.einride.tech/cloudrunner/cloudotel"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Deprecated: use cloudotel.TraceExporterConfig.
type ExporterConfig = cloudotel.TraceExporterConfig

// StartExporter starts the OpenTelemetry Cloud Trace exporter.
// Deprecated: use cloudotel.StartTraceExporter.
func StartExporter(
	ctx context.Context,
	exporterConfig ExporterConfig,
	resource *resource.Resource,
) (func(context.Context) error, error) {
	return cloudotel.StartTraceExporter(ctx, exporterConfig, resource)
}
