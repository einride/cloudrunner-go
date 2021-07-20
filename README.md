# Cloud Runner

Get up and running with [Go][go] and [gRPC][grpc] on [Google Cloud
Platform][gcp], with this lightweight, opinionated, batteries-included
service SDK.

[go]: https://golang.org/
[gcp]: https://cloud.google.com/
[cloud-run]: https://cloud.google.com/run
[grpc]: https://grpc.io

## Features

Run your application with [`cloudrunner.Run`][run], and you get:

- Logging integrated with [Cloud Logging][cloud-logging] using [Zap][zap].
- Tracing integrated with [Cloud Trace][cloud-trace] using
  [OpenTelemetry Go][open-telemetry-go].
- Metrics integrated with [Cloud Monitoring][cloud-monitoring] using
  [OpenTelemetry Go][open-telemetry-go].
- Profiling integrated with [Cloud Profiler][cloud-profiler] using
  the [Google Cloud Go SDK][google-cloud-go].

[run]: ./run.go
[cloud-logging]: https://cloud.google.com/logging
[zap]: go.uber.org/zap
[cloud-trace]: https://cloud.google.com/trace
[open-telemetry-go]: https://go.opentelemetry.io/otel
[cloud-monitoring]: https://cloud.google.com/monitoring
[cloud-profiler]: https://cloud.google.com/profiler
[google-cloud-go]: https://cloud.google.com/go

To help you build gRPC microservices, you also get:

- Server-to-server authentication, client retries, and more for gRPC
  clients with [`cloudrunner.DialService`][dial-service].
- Request logging, tracing, and more, for gRPC servers with
  [`cloudrunner.NewGRPCServer`][grpc-server].

[dial-service]: ./dialservice.go
[grpc-server]: ./grpcserver.go

## Get up and running

Install the package:

```bash
$ go get go.einride.tech/cloudrunner
```

Try out a minimal example:

```go
package main

import (
	"context"
	"log"

	"go.einride.tech/cloudrunner"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	if err := cloudrunner.Run(func(ctx context.Context) error {
		cloudrunner.Logger(ctx).Info("hello world")
		grpcServer := cloudrunner.NewGRPCServer(ctx)
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
		return cloudrunner.ListenGRPC(ctx, grpcServer)
	}); err != nil {
		log.Fatal(err)
	}
}
```

## Configuration

The service is configured with environment variables.

When the service is running [on GCE][on-gce], all built-in integrations
are turned on by default.

[Service-specific config][options] is supported out of the box.

[options]: ./options.go
[on-gce]: https://pkg.go.dev/cloud.google.com/go/compute/metadata#OnGCE

Invoke your service with `-help` to show available configuration.

```
 $ go run go.einride.tech/cloudrunner/examples/cmd/grpc-server -help

Usage of grpc-server:

  -config string
        load environment from a YAML service specification
  -help
        show help then exit
  -validate
        validate config then exit

Runtime configuration of grpc-server:

CONFIG         ENV                                  TYPE                            DEFAULT        ON GCE
cloudrunner    PORT                                 int                             8080
cloudrunner    K_SERVICE                            string
cloudrunner    K_REVISION                           string
cloudrunner    K_CONFIGURATION                      string
cloudrunner    GOOGLE_CLOUD_PROJECT                 string
cloudrunner    SERVICE_ACCOUNT                      string
cloudrunner    SERVICE_VERSION                      string
cloudrunner    LOGGER_DEVELOPMENT                   bool                            true           false
cloudrunner    LOGGER_LEVEL                         zapcore.Level                   debug          info
cloudrunner    LOGGER_REPORTERRORS                  bool                                           true
cloudrunner    PROFILER_ENABLED                     bool                                           true
cloudrunner    PROFILER_MUTEXPROFILING              bool
cloudrunner    PROFILER_ALLOCFORCEGC                bool                            true
cloudrunner    TRACEEXPORTER_ENABLED                bool                                           true
cloudrunner    SERVER_TIMEOUT                       time.Duration                   290s
cloudrunner    SERVER_RECOVERPANICS                 bool                                           true
cloudrunner    CLIENT_TIMEOUT                       time.Duration                   10s
cloudrunner    CLIENT_RETRY_ENABLED                 bool                            true
cloudrunner    CLIENT_RETRY_INITIALBACKOFF          time.Duration                   200ms
cloudrunner    CLIENT_RETRY_MAXBACKOFF              time.Duration                   60s
cloudrunner    CLIENT_RETRY_MAXATTEMPTS             int                             5
cloudrunner    CLIENT_RETRY_BACKOFFMULTIPLIER       float64                         1.3
cloudrunner    CLIENT_RETRY_RETRYABLESTATUSCODES    []codes.Code                    Unavailable
cloudrunner    REQUESTLOGGER_MESSAGESIZELIMIT       int                                            1024
cloudrunner    REQUESTLOGGER_CODETOLEVEL            map[codes.Code]zapcore.Level
cloudrunner    REQUESTLOGGER_STATUSTOLEVEL          map[int]zapcore.Level

Build-time configuration of grpc-server:

LDFLAG                                                     TYPE      VALUE
go.einride.tech/cloudrunner/cloudruntime.serviceVersion    string
```
