package cloudotel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"go.einride.tech/cloudrunner/cloudruntime"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// ResourceConfig configures NewResourceWithConfig behavior.
type ResourceConfig struct {
	// AllowPartialResource allows startup to continue when OTel resource detection is partial (e.g. some detectors
	// fail). A warning is logged.
	AllowPartialResource bool
	// AllowSchemaURLConflict allows startup to continue when OTel resource schema URLs conflict during merge.
	// A warning is logged.
	AllowSchemaURLConflict bool
}

// NewResource creates and detects attributes for a new OpenTelemetry resource.
func NewResource(ctx context.Context) (*resource.Resource, error) {
	return NewResourceWithConfig(ctx, ResourceConfig{})
}

// NewResourceWithConfig creates and detects attributes for a new OpenTelemetry resource, with optional tolerance for
// partial resources and schema URL conflicts configured via ResourceConfig.
func NewResourceWithConfig(ctx context.Context, config ResourceConfig) (*resource.Resource, error) {
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
		if config.AllowPartialResource && errors.Is(err, resource.ErrPartialResource) {
			slog.WarnContext(ctx, "partial otel resource being used", slog.Any("error", err))
			return result, nil
		}
		if config.AllowSchemaURLConflict && errors.Is(err, resource.ErrSchemaURLConflict) {
			slog.WarnContext(ctx, "otel resource schema merge conflict, using anyway", slog.Any("error", err))
			return result, nil
		}
		return nil, fmt.Errorf("init telemetry resource: %w", err)
	}
	return result, nil
}
