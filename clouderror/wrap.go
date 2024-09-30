package clouderror

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"syscall"

	"golang.org/x/net/http2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Wrap masks the gRPC status of the provided error by replacing it with the provided status.
func Wrap(err error, s *status.Status) error {
	return &wrappedStatusError{status: s, err: err, caller: NewCaller(runtime.Caller(1))}
}

// WrapCaller masks the gRPC status of the provided error by replacing it with the provided status.
// The call site of the error is captured from the provided caller.
func WrapCaller(err error, s *status.Status, caller Caller) error {
	return &wrappedStatusError{status: s, err: err, caller: caller}
}

// WrapTransient masks the gRPC status of the provided error by replacing the status message.
// If the original error has transient (retryable) gRPC status code, the status code is forwarded.
// Otherwise, the status code is masked with INTERNAL.
func WrapTransient(err error, msg string) error {
	return WrapTransientCaller(err, msg, NewCaller(runtime.Caller(1)))
}

// WrapTransient masks the gRPC status of the provided error by replacing the status message.
// If the original error has transient (retryable) gRPC status code, the status code is forwarded.
// Otherwise, the status code is masked with INTERNAL.
// The call site of the error is captured from the provided caller.
func WrapTransientCaller(err error, msg string, caller Caller) error {
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
			return &wrappedStatusError{status: status.New(s.Code(), msg), err: err, caller: caller}
		}
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return &wrappedStatusError{status: status.New(codes.DeadlineExceeded, msg), err: err, caller: caller}
	case errors.Is(err, context.Canceled):
		return &wrappedStatusError{status: status.New(codes.Canceled, msg), err: err, caller: caller}
	case errors.Is(err, syscall.ECONNRESET):
		return &wrappedStatusError{status: status.New(codes.Unavailable, msg), err: err, caller: caller}
	case errors.As(err, &http2.GoAwayError{}):
		return &wrappedStatusError{status: status.New(codes.Unavailable, msg), err: err, caller: caller}
	case os.IsTimeout(err):
		return &wrappedStatusError{status: status.New(codes.Unavailable, msg), err: err, caller: caller}
	default:
		return &wrappedStatusError{status: status.New(codes.Internal, msg), err: err, caller: caller}
	}
}

type wrappedStatusError struct {
	status *status.Status
	err    error
	caller Caller
}

// Caller returns the error's caller.
func (w *wrappedStatusError) Caller() (pc uintptr, file string, line int, ok bool) {
	return w.caller.pc, w.caller.file, w.caller.line, w.caller.ok
}

// String implements fmt.Stringer.
func (w *wrappedStatusError) String() string {
	return w.Error()
}

// Error implements error.
func (w *wrappedStatusError) Error() string {
	return fmt.Sprintf("%v: %s: %v", w.status.Code(), w.status.Message(), w.err)
}

// GRPCStatus returns the gRPC status of the wrapped error.
func (w *wrappedStatusError) GRPCStatus() *status.Status {
	return w.status
}

// Unwrap implements error unwrapping.
func (w *wrappedStatusError) Unwrap() error {
	return w.err
}

// Caller is the caller info for an error.
type Caller struct {
	pc   uintptr
	file string
	line int
	ok   bool
}

// NewCaller creates a new caller.
func NewCaller(pc uintptr, file string, line int, ok bool) Caller {
	return Caller{pc: pc, file: file, line: line, ok: ok}
}
