package cloudrunner

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"go.einride.tech/cloudrunner/cloudclient"
	"go.einride.tech/cloudrunner/cloudconfig"
	"go.einride.tech/cloudrunner/cloudotel"
	"go.einride.tech/cloudrunner/cloudprofiler"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.einride.tech/cloudrunner/cloudserver"
	"go.einride.tech/cloudrunner/cloudslog"
	"go.einride.tech/cloudrunner/cloudtrace"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.einride.tech/protobuf-sensitive/protosensitive"
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
	TraceExporter cloudotel.TraceExporterConfig
	// MetricExporter contains metric exporter config.
	MetricExporter cloudotel.MetricExporterConfig
	// Server contains server config.
	Server cloudserver.Config
	// Client contains client config.
	Client cloudclient.Config
	// RequestLogger contains request logging config.
	RequestLogger cloudrequestlog.Config
	// Artificial shutdown delay, allows for the service to
	// process all incoming requests properly, before cancelling
	// the root context.
	// Note: Values higher than 10s will not be respected by cloudrun itself.
	ShutdownDelay time.Duration
}

// Run a service.
// Configuration of the service is loaded from the environment.
func Run(fn func(context.Context) error, options ...Option) (err error) {
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
	if *validate {
		run.configOptions = append(run.configOptions, cloudconfig.WithOptionalSecrets())
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
	if run.traceMiddleware.TraceHook == nil {
		run.traceMiddleware.TraceHook = cloudtrace.IDHook
	}
	run.serverMiddleware.Config = run.config.Server
	run.requestLoggerMiddleware.Config = run.config.RequestLogger
	if run.requestLoggerMiddleware.MessageTransformer == nil {
		run.requestLoggerMiddleware.MessageTransformer = protosensitive.Redact
	}
	ctx = withRunContext(ctx, &run)
	ctx = cloudruntime.WithConfig(ctx, run.config.Runtime)
	logger, err := cloudzap.NewLogger(run.config.Logger)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	// Set the global default log/slog logger to write to our zap logger
	slog.SetDefault(slog.New(cloudslog.NewHandler(run.config.Logger)))
	run.loggerMiddleware.Logger = logger
	ctx = cloudzap.WithLogger(ctx, logger)
	// Set up shutdown delay
	if run.config.ShutdownDelay.Seconds() != 0 {
		sigCtx := ctx
		ctx, cancel = context.WithCancel(context.WithoutCancel(ctx))
		go func() {
			<-sigCtx.Done()
			slog.InfoContext(ctx, "delaying shutdown", slog.Duration("duration", run.config.ShutdownDelay))
			time.Sleep(run.config.ShutdownDelay)
			cancel()
		}()
		defer cancel()
	}
	if err := cloudprofiler.Start(run.config.Profiler); err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	resource, err := cloudotel.NewResource(ctx)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	stopTraceExporter, err := cloudotel.StartTraceExporter(ctx, run.config.TraceExporter, resource)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	defer stopTraceExporter()
	stopMetricExporter, err := cloudotel.StartMetricExporter(ctx, run.config.MetricExporter, resource)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	defer stopMetricExporter()
	cloudotel.RegisterErrorHandler(ctx)
	buildInfo, _ := debug.ReadBuildInfo()
	slog.InfoContext(
		ctx,
		"up and running",
		slog.Any("config", config),
		slog.Any("resource", resource),
		slog.Any("buildInfo", buildInfo),
	)
	defer slog.InfoContext(ctx, "goodbye")
	defer func() {
		if r := recover(); r != nil {
			var errorAttr slog.Attr
			if err2, ok := r.(error); ok {
				errorAttr = slog.Any("error", err2)
				err = err2
			} else {
				errorAttr = slog.Any("error", r)
				err = fmt.Errorf("recovered panic")
			}
			slog.ErrorContext(
				ctx,
				"recovered panic",
				errorAttr,
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()
	return fn(ctx)
}

type runContext struct {
	config                    runConfig
	configOptions             []cloudconfig.Option
	grpcServerOptions         []grpc.ServerOption
	loggerMiddleware          cloudzap.Middleware
	serverMiddleware          cloudserver.Middleware
	clientMiddleware          cloudclient.Middleware
	requestLoggerMiddleware   cloudrequestlog.Middleware
	traceMiddleware           cloudtrace.Middleware
	securityHeadersMiddleware cloudserver.SecurityHeadersMiddleware
}

type runContextKey struct{}

func withRunContext(ctx context.Context, run *runContext) context.Context {
	return context.WithValue(ctx, runContextKey{}, run)
}

func getRunContext(ctx context.Context) (*runContext, bool) {
	result, ok := ctx.Value(runContextKey{}).(*runContext)
	return result, ok
}
