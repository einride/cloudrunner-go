package cloudrequestlog

import (
	"log/slog"
	"math"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func attrToField(attr slog.Attr) zapcore.Field {
	if attr.Equal(slog.Attr{}) {
		// Ignore empty attrs.
		return zap.Skip()
	}
	switch attr.Value.Kind() {
	case slog.KindBool:
		return zap.Bool(attr.Key, attr.Value.Bool())
	case slog.KindDuration:
		return zap.Duration(attr.Key, attr.Value.Duration())
	case slog.KindFloat64:
		return zap.Float64(attr.Key, attr.Value.Float64())
	case slog.KindInt64:
		return zap.Int64(attr.Key, attr.Value.Int64())
	case slog.KindString:
		return zap.String(attr.Key, attr.Value.String())
	case slog.KindTime:
		return zap.Time(attr.Key, attr.Value.Time())
	case slog.KindUint64:
		return zap.Uint64(attr.Key, attr.Value.Uint64())
	case slog.KindGroup:
		if attr.Key == "" {
			// Inlines recursively.
			return zap.Inline(groupObject(attr.Value.Group()))
		}
		return zap.Object(attr.Key, groupObject(attr.Value.Group()))
	case slog.KindLogValuer:
		return attrToField(slog.Attr{
			Key: attr.Key,
			// TODO: resolve the value in a lazy way.
			// This probably needs a new Zap field type
			// that can be resolved lazily.
			Value: attr.Value.Resolve(),
		})
	default:
		return zap.Any(attr.Key, attr.Value.Any())
	}
}

// groupObject holds all the Attrs saved in a slog.GroupValue.
type groupObject []slog.Attr

func (gs groupObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, attr := range gs {
		attrToField(attr).AddTo(enc)
	}
	return nil
}

func fieldToAttr(field zapcore.Field) slog.Attr {
	switch field.Type {
	case zapcore.StringType:
		return slog.String(field.Key, field.String)
	case zapcore.Int64Type:
		return slog.Int64(field.Key, field.Integer)
	case zapcore.Int32Type:
		return slog.Int(field.Key, int(field.Integer))
	case zapcore.Uint64Type:
		return slog.Uint64(field.Key, uint64(field.Integer))
	case zapcore.Float64Type:
		return slog.Float64(field.Key, math.Float64frombits(uint64(field.Integer)))
	case zapcore.BoolType:
		return slog.Bool(field.Key, field.Integer == 1)
	case zapcore.TimeType:
		if field.Interface != nil {
			loc, ok := field.Interface.(*time.Location)
			if ok {
				return slog.Time(field.Key, time.Unix(0, field.Integer).In(loc))
			}
		}
		return slog.Time(field.Key, time.Unix(0, field.Integer))
	case zapcore.DurationType:
		return slog.Duration(field.Key, time.Duration(field.Integer))
	default:
		return slog.Any(field.Key, field.Interface)
	}
}
