module go.opentelemetry.io/otel/example/profile/pprofconvert

go 1.21

replace go.opentelemetry.io/otel => ../../../

replace go.opentelemetry.io/otel/sdk/profile => ./../../../sdk/profile

require go.opentelemetry.io/proto/otlp v1.3.1

require (
	github.com/grafana/pyroscope/api v0.4.0
	go.opentelemetry.io/otel/sdk v1.27.0
)

require google.golang.org/protobuf v1.34.1 // indirect
