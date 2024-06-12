package profile

import "context"

type Exporter[Profile any] interface {
	ExportProfiles(ctx context.Context, profiles []*Profile) error
	Shutdown(ctx context.Context) error
}

type noopExporter[Profile any] struct {
}

func (e noopExporter[Profile]) ExportProfiles(ctx context.Context, profiles []*Profile) error {
	return nil
}

func (e noopExporter[Profile]) Shutdown(ctx context.Context) error {
	return nil
}
