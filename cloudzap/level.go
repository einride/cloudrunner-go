package cloudzap

import (
	"log/slog"

	"go.uber.org/zap/zapcore"
)

// LevelToSeverity converts a zapcore.Level to its corresponding Cloud Logging severity level.
// See: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity.
func LevelToSeverity(l zapcore.Level) string {
	switch l {
	case zapcore.DebugLevel:
		return "DEBUG"
	case zapcore.InfoLevel:
		return "INFO"
	case zapcore.WarnLevel:
		return "WARNING"
	case zapcore.ErrorLevel:
		return "ERROR"
	case zapcore.DPanicLevel:
		return "CRITICAL"
	case zapcore.PanicLevel:
		return "ALERT"
	case zapcore.FatalLevel:
		return "EMERGENCY"
	default:
		return "DEFAULT"
	}
}

// LevelToSlog converts a [zapcore.Level] to an [slog.Level].
func LevelToSlog(l zapcore.Level) slog.Level {
	switch l {
	case zapcore.DebugLevel:
		return slog.LevelDebug
	case zapcore.InfoLevel:
		return slog.LevelInfo
	case zapcore.WarnLevel:
		return slog.LevelWarn
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
