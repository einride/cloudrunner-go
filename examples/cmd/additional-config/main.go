package main

import (
	"context"
	"log"

	"go.einride.tech/cloudrunner"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	var config struct {
		Foo string `default:"bar" onGCE:"baz"`
	}
	if err := cloudrunner.Run(
		func(ctx context.Context) error {
			grpcServer := cloudrunner.NewGRPCServer(ctx)
			healthServer := health.NewServer()
			grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
			return cloudrunner.ListenGRPC(ctx, grpcServer)
		},
		cloudrunner.WithConfig("example", &config),
	); err != nil {
		log.Fatal(err)
	}
}
