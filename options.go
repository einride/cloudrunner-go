package cloudrunner

import (
	"context"

	"go.einride.tech/cloudrunner/cloudconfig"
	"go.einride.tech/cloudrunner/cloudtrace"
	"google.golang.org/grpc"
)

// Option provides optional configuration for a run context.
type Option func(*runContext)

// WithConfig configures an additional config struct to be loaded.
func WithConfig(name string, config interface{}) Option {
	return func(run *runContext) {
		run.configOptions = append(run.configOptions, cloudconfig.WithAdditionalSpec(name, config))
	}
}

// WithOptions configures the run context with a list of options.
func WithOptions(options []Option) Option {
	return func(run *runContext) {
		for _, option := range options {
			option(run)
		}
	}
}

// WithGRPCServerOptions configures the run context with additional default options for NewGRPCServer.
func WithGRPCServerOptions(grpcServerOptions ...grpc.ServerOption) Option {
	return func(run *runContext) {
		run.grpcServerOptions = append(run.grpcServerOptions, grpcServerOptions...)
	}
}

// WithTraceHook configures the run context with a trace hook.
func WithTraceHook(traceHook func(context.Context, cloudtrace.Context) context.Context) Option {
	return func(run *runContext) {
		run.traceMiddleware.TraceHook = traceHook
	}
}
