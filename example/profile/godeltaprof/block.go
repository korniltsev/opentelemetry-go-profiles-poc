package godeltaprof

import (
	"context"
	orig "github.com/grafana/pyroscope-go/godeltaprof"
	otlpgodeltaprof "github.com/grafana/pyroscope-go/godeltaprof/otlp"
	"go.opentelemetry.io/otel/internal/global"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/profile"
	pbcommon "go.opentelemetry.io/proto/otlp/common/v1"
	pb "go.opentelemetry.io/proto/otlp/profiles/v1experimental"
	"math/rand"
	"time"
)

type blockProfiler struct {
	profiler *otlpgodeltaprof.BlockProfiler
	scope    instrumentation.Scope
	name     string
}

func (b *blockProfiler) Profile(ctx context.Context) (*pb.ScopeProfiles, error) {
	global.Debug(b.name + " profile collection")
	st := time.Now()
	p, err := b.profiler.Profile()
	if err != nil {
		return nil, err
	}
	et := time.Now()
	profileID := make([]byte, 16)
	rand.Read(profileID)
	pc := &pb.ProfileContainer{
		ProfileId:         profileID,
		StartTimeUnixNano: uint64(st.UnixNano()),
		EndTimeUnixNano:   uint64(et.UnixNano()),
		Profile:           p,
	}
	sp := &pb.ScopeProfiles{
		Scope: &pbcommon.InstrumentationScope{
			Name:    b.scope.Name,
			Version: b.scope.Version,
		},
		SchemaUrl: b.scope.SchemaURL,
		Profiles:  []*pb.ProfileContainer{pc},
	}
	return sp, nil
}

// todo options
func NewBlockProfiler(scope instrumentation.Scope) profile.Profiler[pb.ScopeProfiles] {
	profiler := otlpgodeltaprof.NewBlockProfilerWithOptions(orig.ProfileOptions{
		GenericsFrames: true,
		LazyMappings:   true,
	})
	return &blockProfiler{
		profiler: profiler,
		scope:    scope,
		name:     "block",
	}
}

// todo options
func NewMutexProfiler(scope instrumentation.Scope) profile.Profiler[pb.ScopeProfiles] {
	profiler := otlpgodeltaprof.NewMutexProfilerWithOptions(orig.ProfileOptions{
		GenericsFrames: true,
		LazyMappings:   true,
	})
	return &blockProfiler{
		profiler: profiler,
		scope:    scope,
		name:     "mutex",
	}
}
