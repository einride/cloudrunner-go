package cloudrunner

import (
	"context"

	"go.einride.tech/cloudrunner/cloudconfig"
	"go.einride.tech/cloudrunner/cloudotel"
	"go.einride.tech/cloudrunner/cloudtrace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Option provides optional configuration for a run context.
type Option func(*runContext)

// WithRequestLoggerMessageTransformer configures the request logger with a message transformer.
// Deprecated: This was historically used for redaction. All proto messages are now automatically redacted.
func WithRequestLoggerMessageTransformer(func(proto.Message) proto.Message) Option {
	return func(*runContext) {}
}

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
// Deprecated: use WithOtelTraceHook instead.
func WithTraceHook(traceHook func(context.Context, cloudtrace.Context) context.Context) Option {
	return func(run *runContext) {
		run.useLegacyTracing = true
		run.traceMiddleware.TraceHook = traceHook
	}
}

// WithTraceHook configures the run context with a trace hook.
func WithOtelTraceHook(traceHook cloudotel.TraceHook) Option {
	return func(run *runContext) {
		run.otelTraceMiddleware.TraceHook = traceHook
	}
}
