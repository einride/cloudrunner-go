package cloudrequestlog

import (
	"context"
	"log/slog"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	fields []zap.Field
	arrays []*arrayField
}

type arrayField struct {
	key    string
	values []any
}

// Add additional fields.
func (m *AdditionalFields) Add(args ...any) {
	m.mu.Lock()
	m.fields = append(m.fields, argsToFieldSlice(args)...)
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

// AppendTo appends the additional fields to the input fields.
func (m *AdditionalFields) AppendTo(fields []zap.Field) []zap.Field {
	m.mu.Lock()
	fields = append(fields, m.fields...)
	for _, array := range m.arrays {
		fields = append(fields, zap.Array(array.key, anyArray(array.values)))
	}
	m.mu.Unlock()
	return fields
}

func (m *AdditionalFields) appendTo(attrs []slog.Attr) []slog.Attr {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, field := range m.fields {
		attrs = append(attrs, fieldToAttr(field))
	}
	for _, array := range m.arrays {
		attrs = append(attrs, slog.Any(array.key, array.values))
	}
	return attrs
}

type anyArray []any

func (oa anyArray) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, o := range oa {
		if err := encoder.AppendReflected(o); err != nil {
			return err
		}
	}
	return nil
}

func argsToFieldSlice(args []any) []zap.Field {
	var attr slog.Attr
	fields := make([]zap.Field, 0, len(args))
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		fields = append(fields, attrToField(attr))
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
	case zapcore.Field:
		return fieldToAttr(x), args[1:]
	case slog.Attr:
		return x, args[1:]
	default:
		return slog.Any(badKey, x), args[1:]
	}
}
