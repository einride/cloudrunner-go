package cloudrunner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// NewGRPCServer creates a new gRPC server preconfigured with middleware for request logging, tracing, etc.
func NewGRPCServer(ctx context.Context, opts ...grpc.ServerOption) *grpc.Server {
	run, ok := getRunContext(ctx)
	if !ok {
		panic("cloudrunner.NewGRPCServer: must be called with a context from cloudrunner.Run")
	}
	unaryTracing := run.otelTraceMiddleware.GRPCServerUnaryInterceptor
	streamTracing := run.otelTraceMiddleware.GRPCStreamServerInterceptor
	if run.useLegacyTracing {
		unaryTracing = run.traceMiddleware.GRPCServerUnaryInterceptor
		streamTracing = run.traceMiddleware.GRPCStreamServerInterceptor
	}
	serverOptions := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			unaryTracing,
			run.requestLoggerMiddleware.GRPCUnaryServerInterceptor, // needs to run after trace
			run.serverMiddleware.GRPCUnaryServerInterceptor,        // needs to run after request logger
		),
		grpc.ChainStreamInterceptor(
			streamTracing,
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
	address := fmt.Sprintf(":%d", run.config.Runtime.Port)
	listener, err := (&net.ListenConfig{}).Listen(
		ctx,
		"tcp",
		address,
	)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		slog.InfoContext(ctx, "gRPCServer shutting down")
		grpcServer.GracefulStop()
	}()
	slog.InfoContext(ctx, "gRPCServer listening", slog.String("address", address))
	return grpcServer.Serve(listener)
}
