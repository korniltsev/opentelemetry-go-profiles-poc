module go.opentelemetry.io/otel/example/profile/grpc

go 1.21

replace go.opentelemetry.io/otel => ../../..

replace go.opentelemetry.io/otel/sdk => ../../../sdk

replace go.opentelemetry.io/otel/sdk/profile => ../../../sdk/profile

replace go.opentelemetry.io/otel/example/profile/godeltaprof => ../godeltaprof

replace go.opentelemetry.io/otel/example/profile/pprofconvert => ../pprofconvert

replace go.opentelemetry.io/otel/exporters/otlp/otlpprofile/otlpprofilegrpc => ../../../exporters/otlp/otlpprofile/otlpprofilegrpc

replace github.com/grafana/pyroscope-go/godeltaprof => github.com/grafana/pyroscope-go/godeltaprof v0.0.0-20240612113540-50232f78fc3f

require (
	go.opentelemetry.io/otel/example/profile/godeltaprof v0.0.0
	go.opentelemetry.io/otel/example/profile/pprofconvert v0.0.0
	go.opentelemetry.io/otel/sdk v1.27.0
	go.opentelemetry.io/otel/sdk/profile v0.0.0
)

require (
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	google.golang.org/grpc v1.64.0 // indirect
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.8-0.20240617054943-50c3bf7462dc // indirect
	github.com/grafana/pyroscope-go/godeltaprof/otlp v0.0.0-20240617054943-50c3bf7462dc // indirect
	github.com/grafana/pyroscope/api v0.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpprofile/otlpprofilegrpc v0.0.0
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)
