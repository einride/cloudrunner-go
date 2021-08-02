package cloudzap

import (
	"fmt"
	"time"

	"go.uber.org/zap/zapcore"
)

// NewEncoderConfig creates a new zapcore.EncoderConfig for structured JSON logging to Cloud Logging.
// See: https://cloud.google.com/logging/docs/agent/logging/configuration#special-fields.
func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:  "time",
		LevelKey: "severity",
		NameKey:  "logger",
		// Omit caller and log structured source location instead.
		// See: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogEntrySourceLocation
		CallerKey:     "",
		MessageKey:    "message",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeTime:    zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: func(duration time.Duration, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(fmt.Sprintf("%gs", duration.Seconds()))
		},
		EncodeLevel: func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(LevelToSeverity(level))
		},
	}
}
