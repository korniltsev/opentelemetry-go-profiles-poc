package profile

import "context"

type simpleProcessor[Profile any] struct {
	exporter Exporter[Profile]
}

func (p *simpleProcessor[Profile]) Process(ctx context.Context, profile *Profile) error {
	return p.exporter.ExportProfiles(ctx, []*Profile{profile})
}

func (p *simpleProcessor[Profile]) Shutdown(ctx context.Context) error {
	return p.exporter.Shutdown(ctx)
}

func (p *simpleProcessor[Profile]) ForceFlush(ctx context.Context) error {
	return nil
}

func NewSimpleProcessor[Profile any](exporter Exporter[Profile]) ProfileProcessor[Profile] {
	return &simpleProcessor[Profile]{exporter: exporter}
}
