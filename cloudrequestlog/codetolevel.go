package cloudrequestlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
)

// CodeToLevel returns the default [zapcore.Level] for requests with the provided [codes.Code].
func CodeToLevel(code codes.Code) zapcore.Level {
	switch code {
	case codes.OK:
		return zap.DebugLevel
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
		codes.ResourceExhausted:
		return zap.WarnLevel
	default:
		return zap.ErrorLevel
	}
}
