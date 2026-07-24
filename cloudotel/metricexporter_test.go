package cloudotel

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

func TestMaskedView(t *testing.T) {
	t.Run("masks configured attributes for matching scope", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithView(maskedView("test.scope", nil, attribute.Key("masked"))),
		)
		meter := provider.Meter("test.scope")
		counter, err := meter.Int64Counter("test.counter")
		assert.NilError(t, err)

		ctx := t.Context()
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("masked", "value"),
			attribute.String("kept", "value"),
		))

		var rm metricdata.ResourceMetrics
		err = reader.Collect(ctx, &rm)
		assert.NilError(t, err)

		attrs := dataPointAttributes(t, rm, "test.counter")
		assert.Assert(t, !attrs.HasValue("masked"))
		assert.Assert(t, attrs.HasValue("kept"))
	})

	t.Run("does not mask attributes for a non-matching scope", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithView(maskedView("other.scope", nil, attribute.Key("masked"))),
		)
		meter := provider.Meter("test.scope")
		counter, err := meter.Int64Counter("test.counter")
		assert.NilError(t, err)

		ctx := t.Context()
		counter.Add(ctx, 1, metric.WithAttributes(attribute.String("masked", "value")))

		var rm metricdata.ResourceMetrics
		err = reader.Collect(ctx, &rm)
		assert.NilError(t, err)

		attrs := dataPointAttributes(t, rm, "test.counter")
		assert.Assert(t, attrs.HasValue("masked"))
	})

	t.Run("drops metrics matching a drop pattern within the scope", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithView(maskedView("test.scope", []string{"test.dropped.*"}, attribute.Key("masked"))),
		)
		meter := provider.Meter("test.scope")
		dropped, err := meter.Int64Counter("test.dropped.counter")
		assert.NilError(t, err)
		kept, err := meter.Int64Counter("test.kept.counter")
		assert.NilError(t, err)

		ctx := t.Context()
		dropped.Add(ctx, 10)
		kept.Add(ctx, 20)

		var rm metricdata.ResourceMetrics
		err = reader.Collect(ctx, &rm)
		assert.NilError(t, err)

		var metricNames []string
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				metricNames = append(metricNames, m.Name)
			}
		}
		assert.DeepEqual(t, metricNames, []string{"test.kept.counter"})
	})
}

func TestMaskedViewAndDropMetricView(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(
			maskedView("test.scope", []string{"test.counter"}, attribute.Key("masked")),
			dropMetricView("test.counter"),
		),
	)
	meter := provider.Meter("test.scope")
	counter, err := meter.Int64Counter("test.counter")
	assert.NilError(t, err)

	ctx := t.Context()
	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("masked", "value")))

	var rm metricdata.ResourceMetrics
	err = reader.Collect(ctx, &rm)
	assert.NilError(t, err)

	var metricNames []string
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			metricNames = append(metricNames, m.Name)
		}
	}
	assert.DeepEqual(t, metricNames, []string(nil))
}

// dataPointAttributes returns the attribute set of the single data point recorded
// for the metric with the given name, failing the test if it is missing or ambiguous.
func dataPointAttributes(t *testing.T, rm metricdata.ResourceMetrics, metricName string) attribute.Set {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != metricName {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok || len(sum.DataPoints) != 1 {
				t.Fatalf("dataPointAttributes(%q) = metric with unexpected data %#v, want a single int64 sum point",
					metricName, m.Data)
			}
			return sum.DataPoints[0].Attributes
		}
	}
	t.Fatalf("dataPointAttributes(%q) = not found", metricName)
	return attribute.Set{}
}
