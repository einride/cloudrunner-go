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
	run, ok := getRunContext(ctx)
	if !ok {
		return nil, fmt.Errorf("cloudrunner.DialService %s: must be called with a context from cloudrunner.Run", target)
	}
	return cloudclient.DialService(
		ctx,
		target,
		append(
			[]grpc.DialOption{
				grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
				grpc.WithDefaultServiceConfig(run.config.Client.AsServiceConfigJSON()),
				grpc.WithChainUnaryInterceptor(
					run.requestLoggerMiddleware.GRPCUnaryClientInterceptor,
					run.clientMiddleware.GRPCUnaryClientInterceptor,
				),
			},
			opts...,
		)...,
	)
}
