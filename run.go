package cloudrunner

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"go.einride.tech/cloudrunner/cloudclient"
	"go.einride.tech/cloudrunner/cloudconfig"
	"go.einride.tech/cloudrunner/cloudotel"
	"go.einride.tech/cloudrunner/cloudprofiler"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudserver"
	"go.einride.tech/cloudrunner/cloudtrace"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// runConfig configures the Run entrypoint from environment variables.
type runConfig struct {
	// Service contains generic service config.
	Service ServiceConfig
	// Logger contains logger config.
	Logger cloudzap.LoggerConfig
	// Profiler contains profiler config.
	Profiler cloudprofiler.Config
	// TraceExporter contains trace exporter config.
	TraceExporter cloudtrace.ExporterConfig
	// Server contains server config.
	Server cloudserver.Config
	// Client contains client config.
	Client cloudclient.Config
	// RequestLogger contains request logging config.
	RequestLogger cloudrequestlog.Config
}

// Run a service.
// Configuration of the service is loaded from the environment.
func Run(run func(context.Context) error, options ...Option) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	usage := flag.Bool("help", false, "show help then exit")
	yamlServiceSpecificationFile := flag.String("config", "", "load environment from a YAML service specification")
	validate := flag.Bool("validate", false, "validate config then exit")
	flag.Parse()
	var runCtx runContext
	for _, option := range options {
		option(&runCtx)
	}
	if *yamlServiceSpecificationFile != "" {
		runCtx.configOptions = append(
			runCtx.configOptions, cloudconfig.WithYAMLServiceSpecificationFile(*yamlServiceSpecificationFile),
		)
	}
	config, err := cloudconfig.New("cloudrunner", &runCtx.runConfig, runCtx.configOptions...)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	if *usage {
		printUsage(flag.CommandLine.Output(), config)
		return nil
	}
	if err := config.Load(); err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	runCtx.runConfig.Service.loadFromRuntime()
	if *validate {
		return nil
	}
	runCtx.traceMiddleware.ProjectID = runCtx.runConfig.Service.ProjectID
	runCtx.serverMiddleware.Config = runCtx.runConfig.Server
	runCtx.requestLoggerMiddleware.Config = runCtx.runConfig.RequestLogger
	ctx = withRunContext(ctx, &runCtx)
	logger, err := cloudzap.NewLogger(runCtx.runConfig.Logger)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	runCtx.loggerMiddleware.Logger = logger
	ctx = cloudzap.WithLogger(ctx, logger)
	if err := cloudprofiler.Start(runCtx.runConfig.Profiler); err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	resource, err := cloudotel.NewResource(ctx)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	stopTraceExporter, err := cloudtrace.StartExporter(ctx, runCtx.runConfig.TraceExporter, resource)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	defer stopTraceExporter()
	logger.Info("up and running", zap.Object("config", config), cloudzap.Resource("resource", resource))
	defer logger.Info("goodbye")
	return run(ctx)
}

type runContext struct {
	runConfig               runConfig
	configOptions           []cloudconfig.Option
	grpcServerOptions       []grpc.ServerOption
	loggerMiddleware        cloudzap.Middleware
	serverMiddleware        cloudserver.Middleware
	clientMiddleware        cloudclient.Middleware
	requestLoggerMiddleware cloudrequestlog.Middleware
	traceMiddleware         cloudtrace.Middleware
}

type runContextKey struct{}

func withRunContext(ctx context.Context, runCtx *runContext) context.Context {
	return context.WithValue(ctx, runContextKey{}, runCtx)
}

func getRunContext(ctx context.Context) (*runContext, bool) {
	result, ok := ctx.Value(runContextKey{}).(*runContext)
	return result, ok
}
