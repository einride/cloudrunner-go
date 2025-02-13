package cloudserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"runtime/debug"

	"go.einride.tech/cloudrunner/clouderror"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudstream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Middleware provides standard middleware for gRPC and HTTP servers.
type Middleware struct {
	// Config for the middleware.
	Config Config
}

// GRPCUnaryServerInterceptor implements grpc.UnaryServerInterceptor.
func (i *Middleware) GRPCUnaryServerInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = clouderror.Wrap(
				fmt.Errorf("recovered panic: %v", r),
				status.New(codes.Internal, "internal error"),
			)
			if additionalFields, ok := cloudrequestlog.GetAdditionalFields(ctx); ok {
				additionalFields.Add(slog.String("stack", string(debug.Stack())))
			}
		}
	}()
	if i.Config.Timeout <= 0 {
		return handler(ctx, req)
	}
	ctx, cancel := context.WithTimeout(ctx, i.Config.Timeout)
	defer cancel()
	resp, err = handler(ctx, req)
	if errors.Is(err, context.DeadlineExceeded) {
		// below call is an inline version of cloudrunner.Wrap in order to avoid circular imports
		return nil, clouderror.WrapCaller(
			err,
			status.New(codes.DeadlineExceeded, "context deadline exceeded"),
			clouderror.NewCaller(runtime.Caller(1)),
		)
	}
	return resp, err
}

// GRPCStreamServerInterceptor implements grpc.StreamServerInterceptor.
func (i *Middleware) GRPCStreamServerInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = clouderror.Wrap(
				fmt.Errorf("recovered panic: %v", r),
				status.New(codes.Internal, "internal error"),
			)
			if additionalFields, ok := cloudrequestlog.GetAdditionalFields(ss.Context()); ok {
				additionalFields.Add(slog.String("stack", string(debug.Stack())))
			}
		}
	}()
	if i.Config.Timeout <= 0 {
		return handler(srv, ss)
	}
	ctx, cancel := context.WithTimeout(ss.Context(), i.Config.Timeout)
	defer cancel()

	if err := handler(srv, cloudstream.NewContextualServerStream(ctx, ss)); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return clouderror.WrapCaller(
				err,
				status.New(codes.DeadlineExceeded, "context deadline exceeded"),
				clouderror.NewCaller(runtime.Caller(1)),
			)
		}
		return err
	}
	return nil
}
