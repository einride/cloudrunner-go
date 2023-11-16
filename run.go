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
	"go.einride.tech/cloudrunner/cloudmonitoring"
	"go.einride.tech/cloudrunner/cloudotel"
	"go.einride.tech/cloudrunner/cloudprofiler"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudruntime"
	"go.einride.tech/cloudrunner/cloudserver"
	"go.einride.tech/cloudrunner/cloudtrace"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.einride.tech/protobuf-sensitive/protosensitive"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// gracefulShutdownMaxGracePeriod is the maximum time we wait for the service to finish calling its cancel function
// after a SIGTERM/SIGINT is sent to us.
// If user is using cloudrunner in a Kubernetes like environment, make sure to set `terminationGracePeriodSeconds`
// (default as 30 seconds) above this value to make sure Kubernetes can wait for enough time for graceful shutdown.
// More info see here:
// https://cloud.google.com/blog/products/containers-kubernetes/kubernetes-best-practices-terminating-with-grace
const gracefulShutdownMaxGracePeriod = time.Second * 10

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
//
// Example usage code can be like:
//
//	  err := cloudrunner.Run(func(ctx context.Context) error {
//			  grpcServer := cloudrunner.NewGRPCServer(ctx)
//			  return cloudrunner.ListenGRPC(ctx, grpcServer)
//		  })
func Run(fn func(context.Context) error, options ...Option) (err error) {
	fnWrapper := func(ctx context.Context, _ *Shutdown) error {
		// Shutdown is not used
		return fn(ctx)
	}
	return RunWithGracefulShutdown(fnWrapper, options...)
}

// RunWithGracefulShutdown runs a service and provides a Shutdown where uer can use to register a
// cancel function that will be called when SIGTERM is received.
// Root context will be canceled after running the registered cancel function.
// If the registered cancel functions runs for time longer than gracefulShutdownMaxGracePeriod, Shutdown
// will move on to shut down root context and exist.
//
// Configuration of the service is loaded from the environment.
//
// Example usage code can be like:
//
//		  err := cloudrunner.RunWithGracefulShutdown(func(ctx context.Context, shutdown *cloudrunner.Shutdown) error {
//				  grpcServer := cloudrunner.NewGRPCServer(ctx)
//				  shutdown.RegisterCancelFunc(func() {
//					  grpcServer.Stop()
//	              // or clean up any other resources here
//				  })
//				  return cloudrunner.ListenGRPC(ctx, grpcServer)
//			  })
func RunWithGracefulShutdown(
	fn func(ctx context.Context, shutdown *Shutdown) error,
	options ...Option,
) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
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
	run.metricMiddleware, err = cloudmonitoring.NewMetricMiddleware()
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	ctx = withRunContext(ctx, &run)
	ctx = cloudruntime.WithConfig(ctx, run.config.Runtime)
	logger, err := cloudzap.NewLogger(run.config.Logger)
	if err != nil {
		return fmt.Errorf("cloudrunner.Run: %w", err)
	}
	// Set the global default log/slog logger to write to our zap logger
	slog.SetDefault(newSlogger(logger))
	run.loggerMiddleware.Logger = logger
	ctx = cloudzap.WithLogger(ctx, logger)
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
	logger.Info(
		"up and running",
		zap.Object("config", config),
		cloudzap.Resource("resource", resource),
		zap.Object("buildInfo", buildInfoMarshaler{buildInfo: buildInfo}),
	)
	defer logger.Info("goodbye")
	defer func() {
		if r := recover(); r != nil {
			var msg zap.Field
			if err2, ok := r.(error); ok {
				msg = zap.Error(err2)
				err = err2
			} else {
				msg = zap.Any("msg", r)
				err = fmt.Errorf("recovered panic")
			}
			logger.Error(
				"recovered panic",
				msg,
				zap.Stack("stack"),
			)
		}
	}()

	shutdown := &Shutdown{
		rootCtxCancel: cancel,
	}
	go shutdown.trapShutdownSignal(ctx, logger)
	return fn(ctx, shutdown)
}

// Shutdown is used for CloudRunner to gracefully shutdown. It makes sure its cancel is called before
// rootCtxCancel is called.
type Shutdown struct {
	rootCtxCancel func()
	cancel        func()
}

// RegisterCancelFunc can register a cancel function which will be called when SIGTERM is received, and it is called
// before Shutdown calling its rootCtxCancel() to cancel the root context.
func (s *Shutdown) RegisterCancelFunc(cancel func()) {
	s.cancel = cancel
}

// trapShutdownSignal blocks and waits for shutdown signal, if received, call s.cancel() then shutdown.
//
//nolint:lll
func (s *Shutdown) trapShutdownSignal(ctx context.Context, logger *zap.Logger) {
	logger.Info("watching for termination signals")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)

	// block and wait for a shutdown signal
	sig := <-sigChan
	logger.Info("got signal:", zap.String("signal", sig.String()))
	if s.cancel == nil {
		logger.Info(
			"cloudrunner.Shutdown is not used. Canceling root context directly. Call RunWithGracefulShutdown(...) to enable graceful shutdown if preferred.",
		)
		s.rootCtxCancel()
		return
	}

	// initiate graceful shutdown by calling s.cancel()
	logger.Info("graceful shutdown has begun")
	gracefulPeriodCtx, gracefulPeriodCtxCancel := context.WithTimeout(ctx, gracefulShutdownMaxGracePeriod)
	go func() {
		s.cancel()
		logger.Info("Shutdown.cancel() has finished, meaning we will shutdown cleanly")
		gracefulPeriodCtxCancel()
	}()

	// block and wait until s.cancel() finish or gracefulPeriodCtx timeout.
	<-gracefulPeriodCtx.Done()
	logger.Info("exiting by canceling root context due to shutdown signal")

	s.rootCtxCancel()
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
	metricMiddleware          cloudmonitoring.MetricMiddleware
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

type buildInfoMarshaler struct {
	buildInfo *debug.BuildInfo
}

func (b buildInfoMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	if b.buildInfo == nil {
		return nil
	}
	encoder.AddString("mainPath", b.buildInfo.Main.Path)
	encoder.AddString("goVersion", b.buildInfo.GoVersion)
	return encoder.AddObject("buildSettings", buildSettingsMarshaler(b.buildInfo.Settings))
}

type buildSettingsMarshaler []debug.BuildSetting

func (b buildSettingsMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	for _, setting := range b {
		encoder.AddString(setting.Key, setting.Value)
	}
	return nil
}

// newSlogger returns a slog logger in which the underlying handler writes to the given zap logger.
// this func is kept here instead of in the cloudslog package to avoid having a api surface
// that encompasses zap in that package.
func newSlogger(zl *zap.Logger) *slog.Logger {
	slogHandler := zapslog.NewHandler(zl.Core(), &zapslog.HandlerOptions{
		AddSource: true, // same as zap's AddCaller
	})
	return slog.New(slogHandler)
}
