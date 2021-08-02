package cloudzap

import "go.uber.org/zap/zapcore"

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
