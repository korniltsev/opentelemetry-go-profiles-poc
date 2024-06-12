package main

import (
	"context"
	"fmt"
	"github.com/go-logr/logr/funcr"
	"go.opentelemetry.io/otel/example/profile/godeltaprof"
	"go.opentelemetry.io/otel/example/profile/pprofconvert"
	"go.opentelemetry.io/otel/exporters/otlp/otlpprofile/otlpprofilegrpc"
	"go.opentelemetry.io/otel/internal/global"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/profile"
	"os"
	"runtime"
	"sync"
	"time"
)

func main() {
	l := funcr.NewJSON(func(obj string) {
		os.Stdout.WriteString(obj + "\n")
	}, funcr.Options{
		Verbosity: 12,
	})
	global.SetLogger(l)

	exporter, err := otlpprofilegrpc.NewExporter()
	if err != nil {
		panic(err)
	}
	processor := profile.NewBatchProfileProcessor(exporter,
		profile.WithMaxExportBatchSize(4),
		profile.WithMaxQueueSize(8),
		profile.WithBatchTimeout(10*time.Second),
	)

	scope := instrumentation.Scope{
		Name:      "example",
		Version:   "0.0.1",
		SchemaURL: "",
	}
	p := profile.NewContinuousScheduler(
		profile.WithPeriodic(10*time.Second, pprofconvert.NewCpuProfiler(scope, 10*time.Second), processor),
		profile.WithPeriodic(10*time.Second, godeltaprof.NewHeapProfiler(scope), processor),
		profile.WithPeriodic(10*time.Second, godeltaprof.NewMutexProfiler(scope), processor),
		profile.WithPeriodic(10*time.Second, godeltaprof.NewBlockProfiler(scope), processor),
	)
	defer func() {
		_ = p.Shutdown(context.TODO())
	}()
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	burn()
}

func burn() {
	for {
		wg := sync.WaitGroup{}
		wg.Add(2)
		go fastFunction(&wg)
		go slowFunction(&wg)
		wg.Wait()
	}
}

//go:noinline
func work(n int) {
	// revive:disable:empty-block this is fine because this is a example app, not real production code
	for i := 0; i < n; i++ {
	}
	fmt.Printf("work\n")
	// revive:enable:empty-block
}

var m sync.Mutex

func fastFunction(wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()

	work(2000000000)

	wg.Done()
}

func slowFunction(wg *sync.WaitGroup) {
	m.Lock()
	defer m.Unlock()

	work(8000000000)
	wg.Done()
}
