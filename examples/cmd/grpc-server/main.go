package main

import (
	"context"
	"log"
	"log/slog"

	"go.einride.tech/cloudrunner"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	if err := cloudrunner.Run(func(ctx context.Context) error {
		slog.InfoContext(ctx, "hello world")
		grpcServer := cloudrunner.NewGRPCServer(ctx)
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
		return cloudrunner.ListenGRPC(ctx, grpcServer)
	}); err != nil {
		log.Fatal(err)
	}
}
