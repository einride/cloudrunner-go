package cloudserver

import (
	"context"
	"fmt"

	"go.einride.tech/cloudrunner/clouderror"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudstream"
	"go.uber.org/zap"
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
				additionalFields.Add(zap.Stack("stack"))
			}
		}
	}()
	if i.Config.Timeout <= 0 {
		return handler(ctx, req)
	}
	ctx, cancel := context.WithTimeout(ctx, i.Config.Timeout)
	defer cancel()
	return handler(ctx, req)
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
				additionalFields.Add(zap.Stack("stack"))
			}
		}
	}()
	if i.Config.Timeout <= 0 {
		return handler(srv, ss)
	}
	ctx, cancel := context.WithTimeout(ss.Context(), i.Config.Timeout)
	defer cancel()

	return handler(srv, cloudstream.NewContextualServerStream(ctx, ss))
}
