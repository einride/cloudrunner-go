package cloudrunner_test

import (
	"context"
	"log"

	"go.einride.tech/cloudrunner"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func ExampleRun_helloWorld() {
	if err := cloudrunner.Run(func(ctx context.Context) error {
		cloudrunner.Logger(ctx).Info("hello world")
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func ExampleRun_gRPCServer() {
	if err := cloudrunner.Run(func(ctx context.Context) error {
		grpcServer := cloudrunner.NewGRPCServer(ctx)
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
		return cloudrunner.ListenGRPC(ctx, grpcServer)
	}); err != nil {
		log.Fatal(err)
	}
}
