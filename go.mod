module go.einride.tech/cloudrunner

go 1.16

require (
	cloud.google.com/go v0.94.1
	cloud.google.com/go/monitoring v0.1.0 // indirect
	cloud.google.com/go/profiler v0.1.0
	cloud.google.com/go/trace v0.1.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.22.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.0.0-RC2
	github.com/google/go-cmp v0.5.6
	go.opentelemetry.io/contrib/detectors/gcp v0.22.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.24.0
	go.opentelemetry.io/contrib/instrumentation/host v0.22.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.22.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.22.0
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/sdk v1.0.1
	go.opentelemetry.io/otel/sdk/metric v0.24.0
	go.uber.org/zap v1.19.1
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	google.golang.org/api v0.56.0
	google.golang.org/genproto v0.0.0-20210831024726-fe130286e0e2
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools/v3 v3.0.3
)
