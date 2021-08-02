package cloudrunner

import (
	"context"
	"fmt"

	"go.einride.tech/cloudrunner/cloudclient"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// DialService dials another service using the default service account's Google ID Token authentication.
func DialService(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	runCtx, ok := getRunContext(ctx)
	if !ok {
		return nil, fmt.Errorf("cloudrunner.DialService %s: must be called with a context from cloudrunner.Run", target)
	}
	return cloudclient.DialService(
		ctx,
		target,
		append(
			[]grpc.DialOption{
				grpc.WithDefaultServiceConfig(runCtx.runConfig.Client.AsServiceConfigJSON()),
				grpc.WithChainUnaryInterceptor(
					otelgrpc.UnaryClientInterceptor(),
					runCtx.requestLoggerMiddleware.GRPCUnaryClientInterceptor,
					runCtx.clientMiddleware.GRPCUnaryClientInterceptor,
				),
				grpc.WithBlock(),
			},
			opts...,
		)...,
	)
}
