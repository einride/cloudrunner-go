package cloudzap

import (
	"go.uber.org/zap" //nolint:gomodguard // cloudzap is a zap integration package
)

const (
	traceKey        = "logging.googleapis.com/trace"
	spanIDKey       = "logging.googleapis.com/spanId"
	traceSampledKey = "logging.googleapis.com/trace_sampled"
)

func Trace(traceID string) zap.Field {
	return zap.String(traceKey, traceID)
}

func SpanID(spanID string) zap.Field {
	return zap.String(spanIDKey, spanID)
}

func TraceSampled(sampled bool) zap.Field {
	return zap.Bool(traceSampledKey, sampled)
}
