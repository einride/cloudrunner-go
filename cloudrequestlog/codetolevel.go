package cloudrequestlog

import (
	"log/slog"

	"google.golang.org/grpc/codes"
)

// CodeToLevel returns the default [slog.Level] for requests with the provided [codes.Code].
func CodeToLevel(code codes.Code) slog.Level {
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
