package cloudzap

import (
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Resource constructs a zap.Field with the given key and OpenTelemetry resource.
func Resource(key string, r *resource.Resource) zap.Field {
	return zap.Object(key, resourceObjectMarshaler{resource: r})
}

type resourceObjectMarshaler struct {
	resource *resource.Resource
}

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (r resourceObjectMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	it := r.resource.Iter()
	for it.Next() {
		attr := it.Attribute()
		if err := encoder.AddReflected(string(attr.Key), attr.Value.AsInterface()); err != nil {
			return err
		}
	}
	return nil
}
