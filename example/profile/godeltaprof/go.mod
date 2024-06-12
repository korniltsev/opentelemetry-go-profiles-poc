module go.opentelemetry.io/otel/example/profile/godeltaprof

go 1.21

replace go.opentelemetry.io/otel => ../../../

replace go.opentelemetry.io/otel/sdk/profile => ./../../../sdk/profile

require (
	github.com/grafana/pyroscope-go/godeltaprof v0.1.8-0.20240617054943-50c3bf7462dc
	github.com/grafana/pyroscope-go/godeltaprof/otlp v0.0.0-20240617054943-50c3bf7462dc
	go.opentelemetry.io/otel v1.27.0
	go.opentelemetry.io/otel/sdk v1.27.0
	go.opentelemetry.io/otel/sdk/profile v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/proto/otlp v1.3.1
)

replace github.com/grafana/pyroscope-go/godeltaprof => github.com/grafana/pyroscope-go/godeltaprof v0.0.0-20240612113540-50232f78fc3f

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)
