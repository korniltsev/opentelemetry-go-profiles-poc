module go.opentelemetry.io/otel/exporters/otlp/otlpprofile/otlpprofilegrpc

go 1.21

replace go.opentelemetry.io/otel => ../../../..

replace go.opentelemetry.io/otel/sdk => ../../../../sdk

replace go.opentelemetry.io/otel/sdk/profile => ../../../../sdk/profile

replace go.opentelemetry.io/otel/sdk/profile/godeltaprof => ./../../../../example/profile/godeltaprof

replace go.opentelemetry.io/otel/sdk/profile/pprofconvert => ./../../../../example/profile/pprofconvert

require go.opentelemetry.io/proto/otlp v1.3.1

require go.opentelemetry.io/otel/sdk/profile v0.0.0

require (
	go.opentelemetry.io/otel v1.27.0
	go.opentelemetry.io/otel/sdk v1.27.0
	google.golang.org/grpc v1.64.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/grafana/pyroscope-go/godeltaprof => github.com/grafana/pyroscope-go/godeltaprof v0.0.0-20240612113540-50232f78fc3f
