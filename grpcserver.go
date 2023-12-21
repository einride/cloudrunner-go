package cloudrunner

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// NewGRPCServer creates a new gRPC server preconfigured with middleware for request logging, tracing, etc.
func NewGRPCServer(ctx context.Context, opts ...grpc.ServerOption) *grpc.Server {
	run, ok := getRunContext(ctx)
	if !ok {
		panic("cloudrunner.NewGRPCServer: must be called with a context from cloudrunner.Run")
	}
	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			//nolint:staticcheck // package is deprecated, replace when possible
			otelgrpc.UnaryServerInterceptor(),
			run.loggerMiddleware.GRPCUnaryServerInterceptor, // adds context logger
			run.traceMiddleware.GRPCServerUnaryInterceptor,  // needs the context logger
			run.metricMiddleware.GRPCUnaryServerInterceptor,
			run.requestLoggerMiddleware.GRPCUnaryServerInterceptor, // needs to run after trace
			run.serverMiddleware.GRPCUnaryServerInterceptor,        // needs to run after request logger
		),
		grpc.ChainStreamInterceptor(
			//nolint:staticcheck // package is deprecated, replace when possible
			otelgrpc.StreamServerInterceptor(),
			run.loggerMiddleware.GRPCStreamServerInterceptor,
			run.traceMiddleware.GRPCStreamServerInterceptor,
			run.metricMiddleware.GRPCStreamServerInterceptor,
			run.requestLoggerMiddleware.GRPCStreamServerInterceptor,
			run.serverMiddleware.GRPCStreamServerInterceptor,
		),
		// For details on keepalive settings, see:
		// https://github.com/grpc/grpc-go/blob/master/Documentation/keepalive.md
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			// If a client pings more than once every 30 seconds, terminate the connection
			MinTime: 30 * time.Second,
			// Allow pings even when there are no active streams
			PermitWithoutStream: true,
		}),
	}
	serverOptions = append(serverOptions, run.grpcServerOptions...)
	serverOptions = append(serverOptions, opts...)
	return grpc.NewServer(serverOptions...)
}

// ListenGRPC binds a listener on the configured port and listens for gRPC requests.
func ListenGRPC(ctx context.Context, grpcServer *grpc.Server) error {
	run, ok := getRunContext(ctx)
	if !ok {
		return fmt.Errorf("cloudrunner.ListenGRPC: must be called with a context from cloudrunner.Run")
	}
	listenerCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	// only after we've already gracefully shut down the server do we cancel the inherent grpcContext
	defer cancel()
	address := fmt.Sprintf(":%d", run.config.Runtime.Port)
	listener, err := (&net.ListenConfig{}).Listen(
		listenerCtx,
		"tcp",
		address,
	)
	if err != nil {
		return err
	}
	go func() {
		// The root context is done, gracefully shutdown the grpcServer. Requests attached to
		// the grpc server are still in flight, as grpcctx is still not canceled.
		<-ctx.Done()
		Logger(ctx).Info("gRPCServer shutting down")
		grpcServer.GracefulStop()
	}()
	Logger(ctx).Info("gRPCServer listening", zap.String("address", address))
	return grpcServer.Serve(listener)
}
