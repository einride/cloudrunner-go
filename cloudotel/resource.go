package cloudotel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
)

// NewResource creates and detects attributes for a new OpenTelemetry resource.
func NewResource(ctx context.Context) (*resource.Resource, error) {
	result, err := resource.New(
		ctx,
		resource.WithTelemetrySDK(),
		resource.WithDetectors(gcp.NewDetector()),
	)
	if err != nil {
		return nil, fmt.Errorf("init telemetry resource: %w", err)
	}
	return result, nil
}
