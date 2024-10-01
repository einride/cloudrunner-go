package cloudslog

import (
	"log/slog"

	"go.opentelemetry.io/otel/sdk/resource"
)

func newResourceValue(r *resource.Resource) resourceValue {
	return resourceValue{Resource: r}
}

type resourceValue struct {
	*resource.Resource
}

func (r resourceValue) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, r.Resource.Len())
	it := r.Resource.Iter()
	for it.Next() {
		attr := it.Attribute()
		attrs = append(attrs, slog.Any(string(attr.Key), attr.Value.AsInterface()))
	}
	return slog.GroupValue(attrs...)
}
