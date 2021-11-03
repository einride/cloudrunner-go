package cloudrunner

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.einride.tech/cloudrunner/cloudclient"
	"go.einride.tech/cloudrunner/cloudconfig"
	"go.einride.tech/cloudrunner/cloudotel"
	"go.einride.tech/cloudrunner/cloudprofiler"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.einride.tech/cloudrunner/cloudserver"
	"go.einride.tech/cloudrunner/cloudtrace"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// runConfig configures the Run entrypoint from environment variables.
type runConfig struct {
	// Runtime contains runtime config.
	Runtime cloudruntime.Config
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
func Run(fn func(context.Context) error, options ...Option) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	usage := flag.Bool("help", false, "show help then exit")
	yamlServiceSpecificationFile := flag.String("config", "", "load environment from a YAML service specification")
	validate := flag.Bool("validate", false, "validate config then exit")
	flag.Parse()
	flag.CommandLine.SetOutput(os.Stdout)
	var run runContext
	for _, option := range options {
		option(&run)
	}
	if *yamlServiceSpecificationFile != "" {
		run.configOptions = append(
			run.configOptions, cloudconfig.WithYAMLServiceSpecificationFile(*yamlServiceSpecificationFile),
		)
	}
	config, err := cloudconfig.New("cloudrunner", &run.config, run.configOptions...)
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
	if err := run.config.Runtime.Autodetect(); err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	if *validate {
		return nil
	}
	run.traceMiddleware.ProjectID = run.config.Runtime.ProjectID
	run.serverMiddleware.Config = run.config.Server
	run.requestLoggerMiddleware.Config = run.config.RequestLogger
	ctx = withRunContext(ctx, &run)
	ctx = cloudruntime.WithConfig(ctx, run.config.Runtime)
	logger, err := cloudzap.NewLogger(run.config.Logger)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	run.loggerMiddleware.Logger = logger
	ctx = cloudzap.WithLogger(ctx, logger)
	if err := cloudprofiler.Start(run.config.Profiler); err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	resource, err := cloudotel.NewResource(ctx)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	stopTraceExporter, err := cloudtrace.StartExporter(ctx, run.config.TraceExporter, resource)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	defer stopTraceExporter()
	logger.Info("up and running", zap.Object("config", config), cloudzap.Resource("resource", resource))
	defer logger.Info("goodbye")
	return fn(ctx)
}

type runContext struct {
	config                  runConfig
	configOptions           []cloudconfig.Option
	grpcServerOptions       []grpc.ServerOption
	loggerMiddleware        cloudzap.Middleware
	serverMiddleware        cloudserver.Middleware
	clientMiddleware        cloudclient.Middleware
	requestLoggerMiddleware cloudrequestlog.Middleware
	traceMiddleware         cloudtrace.Middleware
}

type runContextKey struct{}

func withRunContext(ctx context.Context, run *runContext) context.Context {
	return context.WithValue(ctx, runContextKey{}, run)
}

func getRunContext(ctx context.Context) (*runContext, bool) {
	result, ok := ctx.Value(runContextKey{}).(*runContext)
	return result, ok
}
