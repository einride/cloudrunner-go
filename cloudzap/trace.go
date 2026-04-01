package cloudzap

import (
	"go.uber.org/zap" //nolint:gomodguard // cloudzap is a zap integration package
)

const (
	traceKey        = "logging.googleapis.com/trace"
	spanIDKey       = "logging.googleapis.com/spanId"
	traceSampledKey = "logging.googleapis.com/trace_sampled"
)

// Trace creates a zap field for the Cloud Logging trace field.
//
// Deprecated: Use log/slog with the cloudslog package instead. The cloudslog.Handler
// automatically injects trace fields from the OpenTelemetry span context.
func Trace(traceID string) zap.Field {
	return zap.String(traceKey, traceID)
}

// SpanID creates a zap field for the Cloud Logging span ID field.
//
// Deprecated: Use log/slog with the cloudslog package instead. The cloudslog.Handler
// automatically injects span ID from the OpenTelemetry span context.
func SpanID(spanID string) zap.Field {
	return zap.String(spanIDKey, spanID)
}

// TraceSampled creates a zap field for the Cloud Logging trace sampled field.
//
// Deprecated: Use log/slog with the cloudslog package instead. The cloudslog.Handler
// automatically injects trace sampled from the OpenTelemetry span context.
func TraceSampled(sampled bool) zap.Field {
	return zap.Bool(traceSampledKey, sampled)
}
