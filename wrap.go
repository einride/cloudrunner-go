package cloudrunner

import (
	"runtime"

	"go.einride.tech/cloudrunner/clouderror"
	"google.golang.org/grpc/status"
)

// Wrap an error with a gRPC status.
func Wrap(err error, s *status.Status) error {
	return clouderror.WrapCaller(err, s, clouderror.NewCaller(runtime.Caller(1)))
}

// WrapTransient wraps transient errors (possibly status.Status) with
// appropriate codes.Code. The returned error will always be a status.Status
// with description set to msg.
func WrapTransient(err error, msg string) error {
	return clouderror.WrapTransientCaller(err, msg, clouderror.NewCaller(runtime.Caller(1)))
}
