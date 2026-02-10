package cloudotel

import (
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"gotest.tools/v3/assert"
)

func TestDropMetricView(t *testing.T) {
	t.Run("drops exact metric name", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithView(dropMetricView("test.dropped.counter")),
		)
		meter := provider.Meter("test")

		// Create two counters - one should be dropped, one should not
		droppedCounter, err := meter.Int64Counter("test.dropped.counter")
		assert.NilError(t, err)
		keptCounter, err := meter.Int64Counter("test.kept.counter")
		assert.NilError(t, err)

		// Record values
		ctx := t.Context()
		droppedCounter.Add(ctx, 10)
		keptCounter.Add(ctx, 20)

		// Collect metrics
		var rm metricdata.ResourceMetrics
		err = reader.Collect(ctx, &rm)
		assert.NilError(t, err)

		// Verify only the kept counter is present
		var metricNames []string
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				metricNames = append(metricNames, m.Name)
			}
		}
		assert.DeepEqual(t, metricNames, []string{"test.kept.counter"})
	})

	t.Run("drops metrics matching wildcard pattern", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithView(dropMetricView("test.dropped.*")),
		)
		meter := provider.Meter("test")

		// Create counters - those matching the pattern should be dropped
		dropped1, err := meter.Int64Counter("test.dropped.counter1")
		assert.NilError(t, err)
		dropped2, err := meter.Int64Counter("test.dropped.counter2")
		assert.NilError(t, err)
		kept, err := meter.Int64Counter("test.kept.counter")
		assert.NilError(t, err)

		// Record values
		ctx := t.Context()
		dropped1.Add(ctx, 10)
		dropped2.Add(ctx, 20)
		kept.Add(ctx, 30)

		// Collect metrics
		var rm metricdata.ResourceMetrics
		err = reader.Collect(ctx, &rm)
		assert.NilError(t, err)

		// Verify only the kept counter is present
		var metricNames []string
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				metricNames = append(metricNames, m.Name)
			}
		}
		assert.DeepEqual(t, metricNames, []string{"test.kept.counter"})
	})

	t.Run("multiple drop views", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithView(
				dropMetricView("test.dropped.first"),
				dropMetricView("test.dropped.second"),
			),
		)
		meter := provider.Meter("test")

		// Create counters
		dropped1, err := meter.Int64Counter("test.dropped.first")
		assert.NilError(t, err)
		dropped2, err := meter.Int64Counter("test.dropped.second")
		assert.NilError(t, err)
		kept, err := meter.Int64Counter("test.kept.counter")
		assert.NilError(t, err)

		// Record values
		ctx := t.Context()
		dropped1.Add(ctx, 10)
		dropped2.Add(ctx, 20)
		kept.Add(ctx, 30)

		// Collect metrics
		var rm metricdata.ResourceMetrics
		err = reader.Collect(ctx, &rm)
		assert.NilError(t, err)

		// Verify only the kept counter is present
		var metricNames []string
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				metricNames = append(metricNames, m.Name)
			}
		}
		assert.DeepEqual(t, metricNames, []string{"test.kept.counter"})
	})
}
