package profile

import "context"

// todo Do we need processor at all? should we make it Exporter and BatchExporter?
type ProfileProcessor[Profile any] interface {
	Process(ctx context.Context, profile *Profile) error
	Shutdown(ctx context.Context) error
	ForceFlush(ctx context.Context) error
}
