package cloudrequestlog

import (
	"context"
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
	arrays map[string][]zapcore.ObjectMarshaler
}

// Add additional fields.
func (m *AdditionalFields) Add(fields ...zap.Field) {
	m.mu.Lock()
	m.fields = append(m.fields, fields...)
	m.mu.Unlock()
}

// AddToArray adds additional objects to an array field.
func (m *AdditionalFields) AddToArray(key string, objects ...zapcore.ObjectMarshaler) {
	m.mu.Lock()
	if m.arrays == nil {
		m.arrays = map[string][]zapcore.ObjectMarshaler{}
	}
	m.arrays[key] = append(m.arrays[key], objects...)
	m.mu.Unlock()
}

// AppendTo appends the additional fields to the input fields.
func (m *AdditionalFields) AppendTo(fields []zap.Field) []zap.Field {
	m.mu.Lock()
	fields = append(fields, m.fields...)
	for key, objects := range m.arrays {
		fields = append(fields, zap.Array(key, objectArray(objects)))
	}
	m.mu.Unlock()
	return fields
}

type objectArray []zapcore.ObjectMarshaler

func (oa objectArray) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, o := range oa {
		if err := encoder.AppendObject(o); err != nil {
			return err
		}
	}
	return nil
}
