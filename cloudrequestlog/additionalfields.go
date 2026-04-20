package cloudrequestlog

import (
	"context"
	"log/slog"
	"sync"
)

type additionalFieldsKey struct{}

// GetAdditionalFields returns the current request metadata.
func GetAdditionalFields(ctx context.Context) (*AdditionalFields, bool) {
	md, ok := ctx.Value(additionalFieldsKey{}).(*AdditionalFields)
	return md, ok
}

// WithAdditionalFields initializes metadata for the current request.
func WithAdditionalFields(ctx context.Context) context.Context {
	return context.WithValue(ctx, additionalFieldsKey{}, &AdditionalFields{})
}

// AdditionalFields for a request log message.
type AdditionalFields struct {
	mu     sync.Mutex
	attrs  []slog.Attr
	arrays []*arrayField
}

type arrayField struct {
	key    string
	values []any
}

// Add additional attrs.
func (m *AdditionalFields) Add(args ...any) {
	m.mu.Lock()
	m.attrs = append(m.attrs, argsToAttrSlice(args)...)
	m.mu.Unlock()
}

// AddToArray adds additional objects to an array field.
func (m *AdditionalFields) AddToArray(key string, objects ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var array *arrayField
	for _, needle := range m.arrays {
		if needle.key == key {
			array = needle
			break
		}
	}
	if array == nil {
		array = &arrayField{key: key}
		m.arrays = append(m.arrays, array)
	}
	array.values = append(array.values, objects...)
}

// AppendTo appends the additional attrs to the input attrs.
func (m *AdditionalFields) AppendTo(attrs []slog.Attr) []slog.Attr {
	m.mu.Lock()
	defer m.mu.Unlock()
	attrs = append(attrs, m.attrs...)
	for _, array := range m.arrays {
		attrs = append(attrs, slog.Any(array.key, array.values))
	}
	return attrs
}

func argsToAttrSlice(args []any) []slog.Attr {
	var attr slog.Attr
	fields := make([]slog.Attr, 0, len(args))
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		fields = append(fields, attr)
	}
	return fields
}

// argsToAttr is copied from the slog stdlib.
func argsToAttr(args []any) (slog.Attr, []any) {
	const badKey = "!BADKEY"
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]
	case slog.Attr:
		return x, args[1:]
	default:
		return slog.Any(badKey, x), args[1:]
	}
}
