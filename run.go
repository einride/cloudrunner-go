package cloudrunner

import (
	"context"
	"errors"
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
	run.otelTraceMiddleware = cloudotel.NewTraceMiddleware()
	for _, option := range options {
		option(&run)
	}
	if *yamlServiceSpecificationFile != "" {
		run.configOptions = append(
			run.configOptions, cloudconfig.WithYAMLServiceSpecificationFile(*yamlServiceSpecificationFile),
		)
	}
	if *validate {
		run.configOptions = append(
			run.configOptions,
			cloudconfig.WithOptionalSecrets(),
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
	if run.traceMiddleware.TraceHook == nil {
		run.traceMiddleware.TraceHook = cloudtrace.IDHook
	}
	run.otelTraceMiddleware.ProjectID = run.config.Runtime.ProjectID
	run.otelTraceMiddleware.EnablePubsubTracing = run.config.Runtime.EnablePubsubTracing
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
	// Set the global default log/slog logger.
	slog.SetDefault(slog.New(cloudslog.NewHandler(cloudslog.LoggerConfig{
		ProjectID:             run.config.Runtime.ProjectID,
		Development:           run.config.Logger.Development,
		Level:                 cloudzap.LevelToSlog(run.config.Logger.Level),
		ProtoMessageSizeLimit: run.config.RequestLogger.MessageSizeLimit,
	})))
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
	stopMetricExporter, err := cloudotel.StartMetricExporter(ctx, run.config.MetricExporter, resource)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	cloudotel.RegisterErrorHandler(ctx)
	buildInfo, _ := debug.ReadBuildInfo()
	go func() {
		<-ctx.Done()
		// Cloud Run sends a SIGTERM and allows for 10 seconds before it completely shuts down
		// the instance.
		// See https://cloud.google.com/run/docs/container-contract#instance-shutdown for more details.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		slog.InfoContext(ctx, "shutting down")
		err := errors.Join(
			stopTraceExporter(shutdownCtx),
			stopMetricExporter(shutdownCtx),
		)
		if err != nil {
			slog.WarnContext(ctx, "unable to call shutdown routines", slog.Any("error", err))
		}
	}()
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
			var msg slog.Attr
			if err2, ok := r.(error); ok {
				msg = slog.Any("error", err2)
				err = err2
			} else {
				msg = slog.Any("msg", r)
				err = fmt.Errorf("recovered panic")
			}
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				"recovered panic",
				msg,
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
	useLegacyTracing          bool
	traceMiddleware           cloudtrace.Middleware
	otelTraceMiddleware       cloudotel.TraceMiddleware
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
