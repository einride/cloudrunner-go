package cloudotel

import (
	"context"
	"fmt"
	"time"

	traceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
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
	// configure open telemetry to read trace context from GCP `x-cloud-trace-context` header.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		gcppropagator.CloudTraceFormatPropagator{},
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	opencensus.InstallTraceBridge()

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
