module go.einride.tech/cloudrunner

go 1.16

require (
	cloud.google.com/go v0.88.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v1.0.0-RC1.0.20210727190337-8bcac983167d
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.0.0-RC1.0.20210727190337-8bcac983167d
	github.com/google/go-cmp v0.5.6
	go.opentelemetry.io/contrib/detectors/gcp v0.22.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.22.0
	go.opentelemetry.io/contrib/instrumentation/host v0.22.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.22.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.22.0
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/sdk v1.0.0-RC2
	go.opentelemetry.io/otel/sdk/metric v0.22.0
	go.uber.org/zap v1.18.1
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	google.golang.org/api v0.52.0
	google.golang.org/genproto v0.0.0-20210722135532-667f2b7c528f
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools/v3 v3.0.3
)
