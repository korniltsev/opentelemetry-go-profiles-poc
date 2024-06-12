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

type heapProfiler struct {
	scope    instrumentation.Scope
	profiler *otlpgodeltaprof.HeapProfiler
}

func NewHeapProfiler(scope instrumentation.Scope) profile.Profiler[pb.ScopeProfiles] {
	profiler := otlpgodeltaprof.NewHeapProfilerWithOptions(orig.ProfileOptions{
		GenericsFrames: true,
		LazyMappings:   true,
	})
	return &heapProfiler{
		profiler: profiler,
		scope:    scope,
	}
}

func (h *heapProfiler) Profile(ctx context.Context) (*pb.ScopeProfiles, error) {
	global.Debug("heap profile collection")
	st := time.Now()
	p, err := h.profiler.Profile()
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
	scope := &pbcommon.InstrumentationScope{
		Name:    h.scope.Name,
		Version: h.scope.Version,
	}
	sp := &pb.ScopeProfiles{
		Scope:     scope,
		SchemaUrl: h.scope.SchemaURL,
		Profiles:  []*pb.ProfileContainer{pc},
	}
	return sp, nil
}
