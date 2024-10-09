package cloudrequestlog

import (
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
)

// CodeToLevel returns the default [zapcore.Level] for requests with the provided [codes.Code].
// Deprecated: Zap has been replaced by slog.
func CodeToLevel(code codes.Code) zapcore.Level {
	switch code {
	case codes.OK:
		return zap.InfoLevel
	case
		codes.NotFound,
		codes.InvalidArgument,
		codes.AlreadyExists,
		codes.FailedPrecondition,
		codes.Unauthenticated,
		codes.PermissionDenied,
		codes.DeadlineExceeded,
		codes.OutOfRange,
		codes.Canceled,
		codes.Aborted,
		codes.Unavailable,
		codes.ResourceExhausted,
		codes.Unimplemented:
		return zap.WarnLevel
	default:
		return zap.ErrorLevel
	}
}

// CodeToLevel returns the default [slog.Level] for requests with the provided [codes.Code].
func codeToLevel(code codes.Code) slog.Level {
	switch code {
	case codes.OK:
		return slog.LevelInfo
	case
		codes.NotFound,
		codes.InvalidArgument,
		codes.AlreadyExists,
		codes.FailedPrecondition,
		codes.Unauthenticated,
		codes.PermissionDenied,
		codes.DeadlineExceeded,
		codes.OutOfRange,
		codes.Canceled,
		codes.Aborted,
		codes.Unavailable,
		codes.ResourceExhausted,
		codes.Unimplemented:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}
