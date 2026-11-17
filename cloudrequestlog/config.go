package cloudrequestlog

import (
	"log/slog"

	"google.golang.org/grpc/codes"
)

// Config contains request logging config.
type Config struct {
	// MessageSizeLimit is the maximum size, in bytes, of requests and responses to log.
	// Messages larger than the limit will be truncated.
	// Default value, 0, means that no messages will be truncated.
	MessageSizeLimit int `onGCE:"1024"`
	// CodeToLevel enables overriding the default gRPC code to level conversion.
	CodeToLevel map[codes.Code]slog.Level
	// StatusToLevel enables overriding the default HTTP status code to level conversion.
	StatusToLevel map[int]slog.Level
	// PropagateErrorRootCause enables propagation of the root cause of failed requests between services.
	// The root cause is an opaque fingerprint of the deepest request in a call chain that failed at error level,
	// and is logged as the rootCause field on every error level request log along the chain.
	// It is passed to callers in a gRPC response trailer, so both caller and callee must enable it.
	// The fingerprint does not expose internal details to callers.
	// HTTP server request logs carry the root cause only for downstream gRPC calls made with the request
	// context through GRPCUnaryClientInterceptor.
	//
	// EXPERIMENTAL: This option is experimental and may be changed or removed in the future.
	PropagateErrorRootCause bool
}
