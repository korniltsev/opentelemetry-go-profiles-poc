package profile

import "context"

type Profiler[Profile any] interface {
	Profile(ctx context.Context) (*Profile, error)
}
