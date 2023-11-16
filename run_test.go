package cloudrunner_test

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"go.einride.tech/cloudrunner"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gotest.tools/v3/assert"
)

func Test_Run_helloWorld(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	err := cloudrunner.Run(func(ctx context.Context) error {
		cloudrunner.Logger(ctx).Info("hello world")
		return nil
	})

	assert.NilError(t, err)
}

func Test_Run_gRPCServer(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	err := cloudrunner.Run(func(ctx context.Context) error {
		grpcServer := cloudrunner.NewGRPCServer(ctx)
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

		// For shutdown gRPC server otherwise we get blocked on ListenGRPC
		go func() {
			time.Sleep(time.Second)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()
		return cloudrunner.ListenGRPC(ctx, grpcServer)
	})

	assert.NilError(t, err)
}

func Test_RunWithGracefulShutdown_gRPCServer(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	err := cloudrunner.RunWithGracefulShutdown(func(ctx context.Context, shutdown *cloudrunner.Shutdown) error {
		grpcServer := cloudrunner.NewGRPCServer(ctx)
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
		shutdown.RegisterCancelFunc(func() {
			grpcServer.Stop()
			healthServer.Shutdown()
		})

		// For shutdown gRPC server otherwise we get blocked on ListenGRPC
		go func() {
			time.Sleep(time.Second)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()
		return cloudrunner.ListenGRPC(ctx, grpcServer)
	})

	assert.NilError(t, err)
}

func Test_RunWithGracefulShutdown_helloWorld_ctx_cancel_should_before_clean_up_function_call(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	if err := cloudrunner.RunWithGracefulShutdown(func(ctx context.Context, shutdown *cloudrunner.Shutdown) error {
		wg := sync.WaitGroup{}
		wg.Add(1)
		cleanup := func() {
			var isRootContextDone bool
			select {
			case <-ctx.Done():
				isRootContextDone = true
			default:
				isRootContextDone = false
			}
			assert.Equal(t, isRootContextDone, false)
			wg.Done()
		}

		shutdown.RegisterCancelFunc(cleanup)
		cloudrunner.Logger(ctx).Info("hello world")

		go func() {
			// Simulating seeding a SIGTERM call.
			time.Sleep(time.Second)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()

		wg.Wait()
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
