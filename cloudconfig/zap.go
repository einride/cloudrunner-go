package cloudconfig

import (
	"log/slog"
	"time"

	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
)

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (c *Config) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	for _, configSpec := range c.configSpecs {
		if err := encoder.AddObject(configSpec.name, fieldSpecsMarshaler(configSpec.fieldSpecs)); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(c.configSpecs))
	for _, configSpec := range c.configSpecs {
		attrs = append(attrs, slog.Any(configSpec.name, fieldSpecsValue(configSpec.fieldSpecs)))
	}
	return slog.GroupValue(attrs...)
}

type fieldSpecsValue []fieldSpec

func (fsv fieldSpecsValue) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(fsv))
	for _, fs := range fsv {
		if fs.Secret {
			attrs = append(attrs, slog.String(fs.Key, "<secret>"))
			continue
		}
		attrs = append(attrs, slog.Any(fs.Key, fs.Value.Interface()))
	}
	return slog.GroupValue(attrs...)
}

type fieldSpecsMarshaler []fieldSpec

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (fm fieldSpecsMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	for _, fs := range fm {
		if fs.Secret {
			encoder.AddString(fs.Key, "<secret>")
			continue
		}
		switch value := fs.Value.Interface().(type) {
		case time.Duration:
			encoder.AddDuration(fs.Key, value)
		case []codes.Code:
			if err := encoder.AddArray(fs.Key, codesMarshaler(value)); err != nil {
				return err
			}
		case map[codes.Code]zapcore.Level:
			if len(value) > 0 {
				if err := encoder.AddObject(fs.Key, codeToLevelMarshaler(value)); err != nil {
					return err
				}
			}
		default:
			if err := encoder.AddReflected(fs.Key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

type codeToLevelMarshaler map[codes.Code]zapcore.Level

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (c codeToLevelMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	for code, level := range c {
		encoder.AddString(code.String(), level.String())
	}
	return nil
}

type codesMarshaler []codes.Code

// MarshalLogArray implements zapcore.ArrayMarshaler.
func (c codesMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, code := range c {
		encoder.AppendString(code.String())
	}
	return nil
}
