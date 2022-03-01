Cloud Runner
============

Get up and running with [Go](https://golang.org/) and [gRPC](https://grpc.io) on [Google Cloud Platform](https://cloud.google.com/), with this lightweight, opinionated, batteries-included service SDK.

Features
--------

Run your application with [`cloudrunner.Run`](./run.go), and you get:

-	Logging integrated with [Cloud Logging](https://cloud.google.com/logging) using [Zap](https://go.uber.org/zap).
-	Tracing integrated with [Cloud Trace](https://cloud.google.com/trace) using[OpenTelemetry Go](https://go.opentelemetry.io/otel).
-	Metrics integrated with [Cloud Monitoring](https://cloud.google.com/monitoring) using[OpenTelemetry Go](https://go.opentelemetry.io/otel).
-	Profiling integrated with [Cloud Profiler](https://cloud.google.com/profiler) using the [Google Cloud Go SDK](https://cloud.google.com/go).

To help you build gRPC microservices, you also get:

-	Server-to-server authentication, client retries, and more for gRPC clients with [`cloudrunner.DialService`](./dialservice.go).
-	Request logging, tracing, and more, for gRPC servers with[`cloudrunner.NewGRPCServer`](./grpcserver.go).

Get up and running
------------------

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

Configuration
-------------

The service is configured with environment variables.

When the service is running [on GCE](https://pkg.go.dev/cloud.google.com/go/compute/metadata#OnGCE), all built-in integrations are turned on by default.

[Service-specific config](./options.go) is supported out of the box.

Invoke your service with `-help` to show available configuration.

<!-- BEGIN usage -->

```
Usage of grpc-server:

  -config string
    	load environment from a YAML service specification
  -help
    	show help then exit
  -validate
    	validate config then exit

Runtime configuration of grpc-server:

CONFIG         ENV                                      TYPE                            DEFAULT                ON GCE
cloudrunner    PORT                                     int                             8080                   
cloudrunner    K_SERVICE                                string                                                 
cloudrunner    K_REVISION                               string                                                 
cloudrunner    K_CONFIGURATION                          string                                                 
cloudrunner    GOOGLE_CLOUD_PROJECT                     string                                                 
cloudrunner    RUNTIME_SERVICEACCOUNT                   string                                                 
cloudrunner    SERVICE_VERSION                          string                                                 
cloudrunner    LOGGER_DEVELOPMENT                       bool                            true                   false
cloudrunner    LOGGER_LEVEL                             zapcore.Level                   debug                  info
cloudrunner    LOGGER_REPORTERRORS                      bool                                                   true
cloudrunner    PROFILER_ENABLED                         bool                                                   true
cloudrunner    PROFILER_MUTEXPROFILING                  bool                                                   
cloudrunner    PROFILER_ALLOCFORCEGC                    bool                            true                   
cloudrunner    TRACEEXPORTER_ENABLED                    bool                                                   true
cloudrunner    TRACEEXPORTER_TIMEOUT                    time.Duration                   10s                    
cloudrunner    TRACEEXPORTER_SAMPLEPROBABILITY          float64                         0.01                   
cloudrunner    METRICEXPORTER_ENABLED                   bool                                                   false
cloudrunner    METRICEXPORTER_INTERVAL                  time.Duration                   60s                    
cloudrunner    METRICEXPORTER_RUNTIMEINSTRUMENTATION    bool                                                   true
cloudrunner    METRICEXPORTER_HOSTINSTRUMENTATION       bool                                                   true
cloudrunner    SERVER_TIMEOUT                           time.Duration                   290s                   
cloudrunner    CLIENT_TIMEOUT                           time.Duration                   10s                    
cloudrunner    CLIENT_RETRY_ENABLED                     bool                            true                   
cloudrunner    CLIENT_RETRY_INITIALBACKOFF              time.Duration                   200ms                  
cloudrunner    CLIENT_RETRY_MAXBACKOFF                  time.Duration                   60s                    
cloudrunner    CLIENT_RETRY_MAXATTEMPTS                 int                             5                      
cloudrunner    CLIENT_RETRY_BACKOFFMULTIPLIER           float64                         2                      
cloudrunner    CLIENT_RETRY_RETRYABLESTATUSCODES        []codes.Code                    Unavailable,Unknown    
cloudrunner    REQUESTLOGGER_MESSAGESIZELIMIT           int                                                    1024
cloudrunner    REQUESTLOGGER_CODETOLEVEL                map[codes.Code]zapcore.Level                           
cloudrunner    REQUESTLOGGER_STATUSTOLEVEL              map[int]zapcore.Level                                  

Build-time configuration of grpc-server:

LDFLAG                                                     TYPE      VALUE
go.einride.tech/cloudrunner/cloudruntime.serviceVersion    string
```

<!-- END usage -->
