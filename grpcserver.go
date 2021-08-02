package cloudrunner

import (
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NewGRPCServer creates a new gRPC server preconfigured with middleware for request logging, tracing, etc.
func NewGRPCServer(ctx context.Context, opts ...grpc.ServerOption) *grpc.Server {
	run, ok := getRunContext(ctx)
	if !ok {
		panic("cloudrunner.NewGRPCServer: must be called with a context from cloudrunner.Run")
	}
	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
			run.loggerMiddleware.GRPCUnaryServerInterceptor,        // adds context logger
			run.traceMiddleware.GRPCServerUnaryInterceptor,         // needs the context logger
			run.requestLoggerMiddleware.GRPCUnaryServerInterceptor, // needs to run after trace
			run.serverMiddleware.GRPCUnaryServerInterceptor,        // needs to run after request logger
		),
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
	address := fmt.Sprintf(":%d", run.config.Service.Port)
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
		Logger(ctx).Info("gRPCServer shutting down")
		grpcServer.GracefulStop()
	}()
	Logger(ctx).Info("gRPCServer listening", zap.String("address", address))
	return grpcServer.Serve(listener)
}
