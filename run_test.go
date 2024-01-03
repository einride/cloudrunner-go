package cloudrunner_test

import (
	"context"
	"log"
	"syscall"
	"testing"
	"time"

	"go.einride.tech/cloudrunner"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gotest.tools/v3/assert"
)

func ExampleRun_helloWorld() {
	if err := cloudrunner.Run(func(ctx context.Context) error {
		cloudrunner.Logger(ctx).Info("hello world")
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

func Test_helloWorldShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWNDELAY", "1s")
	if err := cloudrunner.Run(func(ctx context.Context) error {
		cloudrunner.Logger(ctx).Info("hello world")
		beforeKill := time.Now()
		go func() {
			// Simulating seeding a SIGTERM call.
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()
		<-ctx.Done()
		afterKill := time.Now()
		assert.Assert(t, afterKill.Sub(beforeKill).Seconds() > 1.0)
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
