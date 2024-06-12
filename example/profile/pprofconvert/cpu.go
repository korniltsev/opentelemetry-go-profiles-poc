package pprofconvert

import (
	"bytes"
	"context"
	"fmt"
	"go.opentelemetry.io/otel/internal/global"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	pbcommon "go.opentelemetry.io/proto/otlp/common/v1"
	pb "go.opentelemetry.io/proto/otlp/profiles/v1experimental"

	"runtime/pprof"
	"sync"
	"time"
)

type CpuProfiler struct {
	duration time.Duration
	sync     sync.Mutex
	scope    instrumentation.Scope
}

// todo options
func NewCpuProfiler(scope instrumentation.Scope, duration time.Duration) *CpuProfiler {
	return &CpuProfiler{
		duration: duration,
		scope:    scope,
	}
}

func (h *CpuProfiler) SetDuration(duration time.Duration) {
	h.sync.Lock()
	defer h.sync.Unlock()
	h.duration = duration
}

func (h *CpuProfiler) Profile(ctx context.Context) (*pb.ScopeProfiles, error) {
	duration := h.getDuration()
	global.Debug("cpu profile collection", "duration", duration)

	buf := bytes.NewBuffer(nil)

	if err := pprof.StartCPUProfile(buf); err != nil {
		return nil, fmt.Errorf("failed to start CPU profile: %w", err)
	}

	t := time.NewTimer(duration)
	defer t.Stop()
	select {
	case <-ctx.Done():
		pprof.StopCPUProfile()
		return nil, fmt.Errorf("profiling cancelled: %w", ctx.Err())
	case <-t.C:
		pprof.StopCPUProfile()
	}

	oprof, err := otlpProfileFromBytes(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to convert profile: %w", err)
	}
	scope := &pbcommon.InstrumentationScope{
		Name:    h.scope.Name,
		Version: h.scope.Version,
	}
	pc := &pb.ProfileContainer{
		Profile: oprof,
	}
	sp := &pb.ScopeProfiles{
		Profiles: []*pb.ProfileContainer{pc},
		Scope:    scope,
	}
	return sp, err
}

func (h *CpuProfiler) getDuration() time.Duration {
	h.sync.Lock()
	defer h.sync.Unlock()
	return h.duration
}
