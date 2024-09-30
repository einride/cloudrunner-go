package cloudrunner

import (
	"runtime"

	"go.einride.tech/cloudrunner/clouderror"
	"google.golang.org/grpc/status"
)

// Wrap masks the gRPC status of the provided error by replacing it with the provided status.
func Wrap(err error, s *status.Status) error {
	return clouderror.WrapCaller(err, s, clouderror.NewCaller(runtime.Caller(1)))
}

// WrapTransient masks the gRPC status of the provided error by replacing the status message.
// If the original error has transient (retryable) gRPC status code, the status code will be forwarded.
// Otherwise, the status code will be masked with INTERNAL.
func WrapTransient(err error, msg string) error {
	return clouderror.WrapTransientCaller(err, msg, clouderror.NewCaller(runtime.Caller(1)))
}
