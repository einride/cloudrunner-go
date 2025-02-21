package cloudrequestlog

import (
	"log/slog"
	"math"
	"time"

	"go.uber.org/zap/zapcore"
)

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
