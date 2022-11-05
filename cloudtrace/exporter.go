package cloudtrace

import (
	"context"
	"fmt"
	"time"

	traceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"go.einride.tech/cloudrunner/cloudotel"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.einride.tech/cloudrunner/cloudzap"
	octrace "go.opencensus.io/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// ExporterConfig configures the trace exporter.
type ExporterConfig struct {
	Enabled           bool          `onGCE:"true"`
	Timeout           time.Duration `default:"10s"`
	SampleProbability float64       `default:"0.01"`
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
	// collect and export traces instrumented by open census
	tracer := tracerProvider.Tracer("bridge")
	octrace.DefaultTracer = opencensus.NewTracer(tracer)

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
