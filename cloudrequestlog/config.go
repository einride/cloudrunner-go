package cloudrequestlog

import (
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
)

// Config contains request logging config.
type Config struct {
	// MessageSizeLimit is the maximum size, in bytes, of requests and responses to log.
	// Messages large than the limit will be truncated.
	// Default value, 0, means that no messages will be truncated.
	MessageSizeLimit int `onGCE:"1024"`
	// CodeToLevel enables overriding the default gRPC code to level conversion.
	CodeToLevel map[codes.Code]zapcore.Level
	// StatusToLevel enables overriding the default HTTP status code to level conversion.
	StatusToLevel map[int]zapcore.Level
}
