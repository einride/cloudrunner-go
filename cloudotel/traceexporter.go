package cloudotel

import (
	"context"
	"fmt"
	"time"

	traceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TraceExporterConfig configures the trace exporter.
type TraceExporterConfig struct {
	Enabled           bool          `onGCE:"true"`
	Timeout           time.Duration `default:"10s"`
	SampleProbability float64       `default:"0.01"`
}

// StartTraceExporter starts the OpenTelemetry Cloud Trace exporter.
func StartTraceExporter(
	ctx context.Context,
	exporterConfig TraceExporterConfig,
	resource *resource.Resource,
) (func(context.Context) error, error) {
	// configure open telemetry to read trace context from GCP `x-cloud-trace-context` header.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		gcppropagator.CloudTraceFormatPropagator{},
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	if !exporterConfig.Enabled {
		return func(context.Context) error { return nil }, nil
	}
	projectID, ok := cloudruntime.ResolveProjectID(ctx)
	if !ok {
		return nil, fmt.Errorf("start trace exporter: unknown project ID")
	}
	exporter, err := traceexporter.New(
		traceexporter.WithProjectID(projectID),
		traceexporter.WithTimeout(exporterConfig.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("start trace exporter: %w", err)
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(exporterConfig.SampleProbability))),
	)
	otel.SetTracerProvider(tracerProvider)
	opencensus.InstallTraceBridge()
	cleanup := func(ctx context.Context) error {
		if err := tracerProvider.ForceFlush(ctx); err != nil {
			return fmt.Errorf("error shutting down trace exporter: %v", err)
		}
		if err := tracerProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("error shutting down trace exporter: %v", err)
		}
		return nil
	}
	return cleanup, nil
}
