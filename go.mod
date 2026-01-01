module go.einride.tech/cloudrunner

go 1.24.0

toolchain go1.24.1

require (
	cloud.google.com/go/compute/metadata v0.9.0
	cloud.google.com/go/logging v1.13.1
	cloud.google.com/go/profiler v0.4.3
	cloud.google.com/go/pubsub/v2 v2.3.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.54.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.30.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator v0.54.0
	github.com/google/go-cmp v0.7.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/soheilhy/cmux v0.1.5
	go.einride.tech/protobuf-sensitive v0.9.0
	go.opentelemetry.io/contrib/detectors/gcp v1.38.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0
	go.opentelemetry.io/contrib/instrumentation/host v0.63.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.63.0
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/bridge/opencensus v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/sdk/metric v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	go.uber.org/zap v1.27.1
	golang.org/x/net v0.48.0
	golang.org/x/oauth2 v0.34.0
	golang.org/x/sync v0.19.0
	google.golang.org/api v0.258.0
	google.golang.org/genproto v0.0.0-20250922171735-9219d122eba9
	google.golang.org/grpc v1.78.0
	google.golang.org/grpc/examples v0.0.0-20240927220217-941102b7811f
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.5.2
)

require (
	cloud.google.com/go v0.121.6 // indirect
	cloud.google.com/go/auth v0.17.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/longrunning v0.6.7 // indirect
	cloud.google.com/go/monitoring v1.24.2 // indirect
	cloud.google.com/go/trace v1.11.6 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.30.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.54.0 // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/pprof v0.0.0-20250602020802-c6617b811d0e // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.7 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250827001030-24949be3fa54 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shirou/gopsutil/v4 v4.25.7 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251213004720-97cd9d5aeac2 // indirect
)

retract (
	v0.77.0 // request logging bug
	v0.75.0 // slog migration bug
)
