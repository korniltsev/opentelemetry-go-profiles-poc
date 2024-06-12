package profile

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"sync"
	"time"
)

type ContinuousScheduler struct {
	jobs   []func()
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

type ContinuousSchedulerOption func(*ContinuousScheduler)

// todo other options?
func WithPeriodic[Profile any](freq time.Duration, profiler Profiler[Profile], processor ProfileProcessor[Profile]) ContinuousSchedulerOption {
	return func(c *ContinuousScheduler) {
		c.jobs = append(c.jobs, func() {
			periodic(c.ctx, freq, profiler, processor)
		})
	}
}

func NewContinuousScheduler(opts ...ContinuousSchedulerOption) *ContinuousScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	c := &ContinuousScheduler{
		ctx:    ctx,
		cancel: cancel,
	}
	for _, opt := range opts {
		opt(c)
	}
	for _, job := range c.jobs {
		c.wg.Add(1)
		go func(job func()) {
			defer c.wg.Done()
			job()
		}(job)
	}
	return c
}

func (c *ContinuousScheduler) Shutdown(ctx context.Context) error {
	c.cancel()
	c.wg.Wait() //todo handle ctx.Done()
	return nil
}

func periodic[Profile any](ctx context.Context, freq time.Duration, profiler Profiler[Profile], processor ProfileProcessor[Profile]) {
	work := func() {
		profile, err := profiler.Profile(ctx)
		if err != nil {
			otel.Handle(fmt.Errorf("ContinuousScheduler failed to profile: %w", err))
			return
		}
		err = processor.Process(ctx, profile)
		if err != nil {
			otel.Handle(fmt.Errorf("ContinuousScheduler failed to process profile: %w", err))
			return
		}
	}
	timer := time.NewTicker(freq)
	defer timer.Stop()
	work()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			work()
		}
	}
}
