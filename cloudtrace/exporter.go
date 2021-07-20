package cloudtrace

import (
	"context"
	"fmt"

	traceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.einride.tech/cloudrunner/cloudotel"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// ExporterConfig configures the trace exporter.
type ExporterConfig struct {
	Enabled bool `onGCE:"true"`
}

// StartExporter starts the OpenTelemetry Cloud Trace exporter.
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
		return nil, fmt.Errorf("start trace exporter: unknown project ID")
	}
	logger, ok := cloudzap.GetLogger(ctx)
	if !ok {
		return nil, fmt.Errorf("start trace exporter: no logger in context")
	}
	exporter, err := traceexporter.New(
		traceexporter.WithProjectID(projectID),
		traceexporter.WithErrorHandler(cloudotel.NewErrorLogger(logger, zap.WarnLevel, "trace exporter error")),
	)
	if err != nil {
		return nil, fmt.Errorf("start trace exporter: %w", err)
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tracerProvider)
	cleanup := func() {
		if err := tracerProvider.ForceFlush(context.Background()); err != nil {
			logger.Error("error shutting down trace exporter", zap.Error(err))
		}
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			logger.Error("error shutting down trace exporter", zap.Error(err))
		}
	}
	return cleanup, nil
}
