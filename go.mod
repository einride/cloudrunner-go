module go.einride.tech/cloudrunner

go 1.26.6

require (
	cloud.google.com/go/compute/metadata v0.9.1
	cloud.google.com/go/logging v1.19.1
	cloud.google.com/go/profiler v0.6.0
	cloud.google.com/go/pubsub/v2 v2.7.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.62.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.38.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator v0.62.0
	github.com/google/go-cmp v0.7.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/soheilhy/cmux v0.1.5
	go.einride.tech/protobuf-sensitive v0.9.0
	go.opentelemetry.io/contrib/detectors/gcp v1.46.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.71.0
	go.opentelemetry.io/contrib/instrumentation/host v0.71.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.71.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.71.0
	go.opentelemetry.io/otel v1.46.0
	go.opentelemetry.io/otel/bridge/opencensus v1.46.0
	go.opentelemetry.io/otel/metric v1.46.0
	go.opentelemetry.io/otel/sdk v1.46.0
	go.opentelemetry.io/otel/sdk/metric v1.46.0
	go.opentelemetry.io/otel/trace v1.46.0
	go.uber.org/zap v1.28.0
	golang.org/x/net v0.59.0
	golang.org/x/oauth2 v0.37.0
	golang.org/x/sync v0.23.0
	google.golang.org/api v0.299.0
	google.golang.org/genproto v0.0.0-20260715232425-e75dac1f907d
	google.golang.org/grpc v1.84.0
	google.golang.org/grpc/examples v0.0.0-20240927220217-941102b7811f
	google.golang.org/protobuf v1.36.12
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.5.2
)

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.23.3 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/iam v1.12.0 // indirect
	cloud.google.com/go/longrunning v1.2.0 // indirect
	cloud.google.com/go/monitoring v1.30.0 // indirect
	cloud.google.com/go/trace v1.16.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.36.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.62.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/ebitengine/purego v0.10.2 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	github.com/google/s2a-go v0.1.10 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.22 // indirect
	github.com/googleapis/gax-go/v2 v2.24.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20260802145828-341c2f0c90b5 // indirect
	github.com/power-devops/perfstat v0.0.0-20260805114148-88456608a4f6 // indirect
	github.com/shirou/gopsutil/v4 v4.26.7 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260715232425-e75dac1f907d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260921155816-b14227669459 // indirect
)

retract (
	v0.77.0 // request logging bug
	v0.75.0 // slog migration bug
)
