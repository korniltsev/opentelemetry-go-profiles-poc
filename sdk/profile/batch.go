package profile

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/internal/global"
)

// Defaults for BatchProfileProcessorOptions.
const (
	DefaultMaxQueueSize       = 2048
	DefaultScheduleDelay      = 5000
	DefaultExportTimeout      = 30000
	DefaultMaxExportBatchSize = 512
)

type BatchProfileProcessorOption func(o *BatchProfileProcessorOptions)

type BatchProfileProcessorOptions struct {
	// MaxQueueSize is the maximum queue size to buffer profiles for delayed processing. If the
	// queue gets full it drops the profiles. Use BlockOnQueueFull to change this behavior.
	// The default value of MaxQueueSize is 2048.
	MaxQueueSize int

	// BatchTimeout is the maximum duration for constructing a batch. ProfileProcessor
	// forcefully sends available profiles when timeout is reached.
	// The default value of BatchTimeout is 5000 msec.
	BatchTimeout time.Duration

	// ExportTimeout specifies the maximum duration for exporting profiles. If the timeout
	// is reached, the export will be cancelled.
	// The default value of ExportTimeout is 30000 msec.
	ExportTimeout time.Duration

	// MaxExportBatchSize is the maximum number of profiles to process in a single batch.
	// If there are more than one batch worth of profiles then it processes multiple batches
	// of profiles one batch after the other without any delay.
	// The default value of MaxExportBatchSize is 512.
	MaxExportBatchSize int

	// BlockOnQueueFull blocks onEnd() and onStart() method if the queue is full
	// AND if BlockOnQueueFull is set to true.
	// Blocking option should be used carefully as it can severely affect the performance of an
	// application.
	BlockOnQueueFull bool
}

// batchProfileProcessor is a SpanProcessor that batches asynchronously-received
// profiles and sends them to a trace.Exporter when complete.
//
// note: This is copied from batch span processor and adapted for profiles for initial POC
type batchProfileProcessor[Profile any] struct {
	e Exporter[Profile]
	o BatchProfileProcessorOptions

	queue   chan *Profile
	dropped uint32

	batch      []*Profile
	batchMutex sync.Mutex
	timer      *time.Timer
	stopWait   sync.WaitGroup
	stopOnce   sync.Once
	stopCh     chan struct{}
	stopped    atomic.Bool
}

var _ ProfileProcessor[any] = (*batchProfileProcessor[any])(nil)

// NewBatchProfileProcessor creates a new ProfileProcessor that will send completed
// profile batches to the exporter with the supplied options.
//
// If the exporter is nil, the profile processor will perform no action.
//
// note: This is copied from batch span processor and adapted for profiles for initial POC
func NewBatchProfileProcessor[Profile any](exporter Exporter[Profile], options ...BatchProfileProcessorOption) ProfileProcessor[Profile] {
	//maxQueueSize := env.BatchSpanProcessorMaxQueueSize(DefaultMaxQueueSize)
	//maxExportBatchSize := env.BatchSpanProcessorMaxExportBatchSize(DefaultMaxExportBatchSize)
	maxQueueSize := DefaultMaxQueueSize
	maxExportBatchSize := DefaultMaxExportBatchSize

	if maxExportBatchSize > maxQueueSize {
		if DefaultMaxExportBatchSize > maxQueueSize {
			maxExportBatchSize = maxQueueSize
		} else {
			maxExportBatchSize = DefaultMaxExportBatchSize
		}
	}

	o := BatchProfileProcessorOptions{
		//BatchTimeout:       time.Duration(env.BatchSpanProcessorScheduleDelay(DefaultScheduleDelay)) * time.Millisecond,
		//ExportTimeout:      time.Duration(env.BatchSpanProcessorExportTimeout(DefaultExportTimeout)) * time.Millisecond,
		BatchTimeout:       time.Duration(DefaultScheduleDelay) * time.Millisecond,
		ExportTimeout:      time.Duration(DefaultExportTimeout) * time.Millisecond,
		MaxQueueSize:       maxQueueSize,
		MaxExportBatchSize: maxExportBatchSize,
	}
	for _, opt := range options {
		opt(&o)
	}
	bsp := &batchProfileProcessor[Profile]{
		e:      exporter,
		o:      o,
		batch:  make([]*Profile, 0, o.MaxExportBatchSize),
		timer:  time.NewTimer(o.BatchTimeout),
		queue:  make(chan *Profile, o.MaxQueueSize),
		stopCh: make(chan struct{}),
	}

	bsp.stopWait.Add(1)
	go func() {
		defer bsp.stopWait.Done()
		bsp.processQueue()
		bsp.drainQueue()
	}()

	return bsp
}

func (bsp *batchProfileProcessor[Profile]) Process(ctx context.Context, profile *Profile) error {
	// Do not enqueue profiles after Shutdown.
	if bsp.stopped.Load() {
		return nil
	}

	// Do not enqueue profiles if we are just going to drop them.
	if bsp.e == nil {
		return nil
	}
	if bsp.o.BlockOnQueueFull {
		bsp.enqueueBlockOnQueueFull(ctx, profile)
	} else {
		bsp.enqueueDrop(profile)
	}
	return nil
}

// Shutdown flushes the queue and waits until all profiles are processed.
// It only executes once. Subsequent call does nothing.
func (bsp *batchProfileProcessor[Profile]) Shutdown(ctx context.Context) error {
	var err error
	bsp.stopOnce.Do(func() {
		bsp.stopped.Store(true)
		wait := make(chan struct{})
		go func() {
			close(bsp.stopCh)
			bsp.stopWait.Wait()
			if bsp.e != nil {
				if err := bsp.e.Shutdown(ctx); err != nil {
					otel.Handle(err)
				}
			}
			close(wait)
		}()
		// Wait until the wait group is done or the context is cancelled
		select {
		case <-wait:
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return err
}

//type forceFlushSpan struct {
//	ReadOnlySpan
//	flushed chan struct{}
//}
//
//func (f forceFlushSpan) SpanContext() trace.SpanContext {
//	return trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: trace.FlagsSampled})
//}

// ForceFlush exports all profiles that have not yet been exported.
func (bsp *batchProfileProcessor[Profile]) ForceFlush(ctx context.Context) error {
	//// Interrupt if context is already canceled.
	//if err := ctx.Err(); err != nil {
	//	return err
	//}
	//
	//// Do nothing after Shutdown.
	//if bsp.stopped.Load() {
	//	return nil
	//}
	//
	//var err error
	//if bsp.e != nil {
	//	flushCh := make(chan struct{})
	//	if bsp.enqueueBlockOnQueueFull(ctx, forceFlushSpan{flushed: flushCh}) {
	//		select {
	//		case <-bsp.stopCh:
	//			// The batchProfileProcessor is Shutdown.
	//			return nil
	//		case <-flushCh:
	//			// Processed any items in queue prior to ForceFlush being called
	//		case <-ctx.Done():
	//			return ctx.Err()
	//		}
	//	}
	//
	//	wait := make(chan error)
	//	go func() {
	//		wait <- bsp.export(ctx)
	//		close(wait)
	//	}()
	//	// Wait until the export is finished or the context is cancelled/timed out
	//	select {
	//	case err = <-wait:
	//	case <-ctx.Done():
	//		err = ctx.Err()
	//	}
	//}
	//return err
	return fmt.Errorf("TODO implement me")
}

// WithMaxQueueSize returns a BatchProfileProcessorOption that configures the
// maximum queue size allowed for a BatchProfileProcessor.
func WithMaxQueueSize(size int) BatchProfileProcessorOption {
	return func(o *BatchProfileProcessorOptions) {
		o.MaxQueueSize = size
	}
}

// WithMaxExportBatchSize returns a BatchProfileProcessorOption that configures
// the maximum export batch size allowed for a BatchProfileProcessor.
func WithMaxExportBatchSize(size int) BatchProfileProcessorOption {
	return func(o *BatchProfileProcessorOptions) {
		o.MaxExportBatchSize = size
	}
}

// WithBatchTimeout returns a BatchProfileProcessorOption that configures the
// maximum delay allowed for a BatchProfileProcessor before it will export any
// held span (whether the queue is full or not).
func WithBatchTimeout(delay time.Duration) BatchProfileProcessorOption {
	return func(o *BatchProfileProcessorOptions) {
		o.BatchTimeout = delay
	}
}

// WithExportTimeout returns a BatchProfileProcessorOption that configures the
// amount of time a BatchProfileProcessor waits for an exporter to export before
// abandoning the export.
func WithExportTimeout(timeout time.Duration) BatchProfileProcessorOption {
	return func(o *BatchProfileProcessorOptions) {
		o.ExportTimeout = timeout
	}
}

// WithBlocking returns a BatchProfileProcessorOption that configures a
// BatchProfileProcessor to wait for enqueue operations to succeed instead of
// dropping data when the queue is full.
func WithBlocking() BatchProfileProcessorOption {
	return func(o *BatchProfileProcessorOptions) {
		o.BlockOnQueueFull = true
	}
}

// export is a subroutine of processing and draining the queue.
func (bsp *batchProfileProcessor[Profile]) export(ctx context.Context) error {
	bsp.timer.Reset(bsp.o.BatchTimeout)

	bsp.batchMutex.Lock()
	defer bsp.batchMutex.Unlock()

	if bsp.o.ExportTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, bsp.o.ExportTimeout)
		defer cancel()
	}

	if l := len(bsp.batch); l > 0 {
		global.Debug("exporting profiles", "count", len(bsp.batch), "total_dropped", atomic.LoadUint32(&bsp.dropped))
		err := bsp.e.ExportProfiles(ctx, bsp.batch)

		// A new batch is always created after exporting, even if the batch failed to be exported.
		//
		// It is up to the exporter to implement any type of retry logic if a batch is failing
		// to be exported, since it is specific to the protocol and backend being sent to.
		bsp.batch = bsp.batch[:0] //todo clear slice

		if err != nil {
			return err
		}
	}
	return nil
}

// processQueue removes profiles from the `queue` channel until processor
// is shut down. It calls the exporter in batches of up to MaxExportBatchSize
// waiting up to BatchTimeout to form a batch.
func (bsp *batchProfileProcessor[Profile]) processQueue() {
	defer bsp.timer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case <-bsp.stopCh:
			return
		case <-bsp.timer.C:
			if err := bsp.export(ctx); err != nil {
				otel.Handle(err)
			}
		case sd := <-bsp.queue:
			//if ffs, ok := sd.(forceFlushSpan); ok {
			//	close(ffs.flushed)
			//	continue
			//}
			bsp.batchMutex.Lock()
			bsp.batch = append(bsp.batch, sd)
			shouldExport := len(bsp.batch) >= bsp.o.MaxExportBatchSize
			bsp.batchMutex.Unlock()
			if shouldExport {
				if !bsp.timer.Stop() {
					<-bsp.timer.C
				}
				if err := bsp.export(ctx); err != nil {
					otel.Handle(err)
				}
			}
		}
	}
}

// drainQueue awaits the any caller that had added to bsp.stopWait
// to finish the enqueue, then exports the final batch.
func (bsp *batchProfileProcessor[Profile]) drainQueue() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case sd := <-bsp.queue:
			//if _, ok := sd.(forceFlushSpan); ok {
			//	// Ignore flush requests as they are not valid spans.
			//	continue
			//}

			bsp.batchMutex.Lock()
			bsp.batch = append(bsp.batch, sd)
			shouldExport := len(bsp.batch) == bsp.o.MaxExportBatchSize
			bsp.batchMutex.Unlock()

			if shouldExport {
				if err := bsp.export(ctx); err != nil {
					otel.Handle(err)
				}
			}
		default:
			// There are no more enqueued profiles. Make final export.
			if err := bsp.export(ctx); err != nil {
				otel.Handle(err)
			}
			return
		}
	}
}

func (bsp *batchProfileProcessor[Profile]) enqueueBlockOnQueueFull(ctx context.Context, sd *Profile) bool {
	select {
	case bsp.queue <- sd:
		return true
	case <-ctx.Done():
		return false
	}
}

func (bsp *batchProfileProcessor[Profile]) enqueueDrop(sd *Profile) bool {
	select {
	case bsp.queue <- sd:
		return true
	default:
		atomic.AddUint32(&bsp.dropped, 1)
	}
	return false
}

// MarshalLog is the marshaling function used by the logging system to represent this Span ProfileProcessor.
func (bsp *batchProfileProcessor[Profile]) MarshalLog() interface{} {
	return struct {
		Type         string
		SpanExporter Exporter[Profile]
		Config       BatchProfileProcessorOptions
	}{
		Type:         "BatchProfileProcessor",
		SpanExporter: bsp.e,
		Config:       bsp.o,
	}
}
