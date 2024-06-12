package otlpprofilegrpc

import (
	"context"
	"go.opentelemetry.io/otel/internal/global"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/profile"
	pbcollector "go.opentelemetry.io/proto/otlp/collector/profiles/v1experimental"
	pbcommon "go.opentelemetry.io/proto/otlp/common/v1"
	pbprofiles "go.opentelemetry.io/proto/otlp/profiles/v1experimental"
	pbresource "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type pbExporter struct {
	scope instrumentation.Scope //todo this should not be here
	svc   pbcollector.ProfilesServiceClient
}

func (e *pbExporter) ExportProfiles(ctx context.Context, profiles []*pbprofiles.ScopeProfiles) error {
	global.Debug("export profiles ", "len", len(profiles))

	req := &pbcollector.ExportProfilesServiceRequest{
		ResourceProfiles: []*pbprofiles.ResourceProfiles{
			{
				Resource: &pbresource.Resource{
					Attributes: []*pbcommon.KeyValue{},
				},
				ScopeProfiles: profiles,
			},
		},
	}
	res, err := e.svc.Export(ctx, req)
	if err != nil {
		return err
	}
	if res.PartialSuccess != nil {
		global.Debug("export profiles response",
			"RejectedProfiles", res.PartialSuccess.RejectedProfiles,
			"ErrorMessage", res.PartialSuccess.ErrorMessage)
	}
	return nil
}

func (e *pbExporter) Shutdown(ctx context.Context) error {
	panic("TODO")
	return nil
}

// todo options
func NewExporter() (profile.Exporter[pbprofiles.ScopeProfiles], error) {
	con, err := grpc.NewClient("localhost:9095", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	svc := pbcollector.NewProfilesServiceClient(con)
	return &pbExporter{
		svc: svc,
	}, nil
}
