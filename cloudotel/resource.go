package cloudotel

import (
	"context"
	"fmt"
	"strconv"

	"go.einride.tech/cloudrunner/cloudruntime"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// NewResource creates and detects attributes for a new OpenTelemetry resource.
func NewResource(ctx context.Context) (*resource.Resource, error) {
	opts := []resource.Option{
		resource.WithTelemetrySDK(),
		resource.WithDetectors(gcp.NewDetector()),
	}

	// In Cloud Run Job, Opentelemetry uses the underlying metadata instance id as the `task_id` label causing a new
	// time series every time the job is run which leads to a high cardinality value so we override it.
	// TODO: Follow-up on https://github.com/GoogleCloudPlatform/opentelemetry-operations-go/issues/874 for possible
	// changes on this.
	if e, ok := cloudruntime.TaskIndex(); ok {
		opts = append(opts, resource.WithAttributes(semconv.FaaSInstanceKey.String(strconv.Itoa(e))))
	}
	if e, ok := cloudruntime.Service(); ok {
		opts = append(opts, resource.WithAttributes(semconv.ServiceName(e)))
	}
	result, err := resource.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("init telemetry resource: %w", err)
	}
	return result, nil
}
