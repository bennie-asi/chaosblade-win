package exec

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// CPURunner drives CPU-bound work across a fixed number of cores until the context ends.
type CPURunner struct {
	cores    int
	percent  int
	duration time.Duration
}

// NewCPURunner constructs a CPURunner with the requested core count and utilization percent.
func NewCPURunner(cores, percent int, duration time.Duration) *CPURunner {
	if cores < 1 {
		cores = 1
	}
	if percent < 1 {
		percent = 1
	}
	if percent > 100 {
		percent = 100
	}
	return &CPURunner{cores: cores, percent: percent, duration: duration}
}

// Run spins CPU-bound goroutines and blocks until the context is canceled.
func (r *CPURunner) Run(ctx context.Context) error {
	if r.duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.duration)
		defer cancel()
	}

	runtime.GOMAXPROCS(r.cores)

	var wg sync.WaitGroup
	wg.Add(r.cores)

	period := 100 * time.Millisecond
	busy := time.Duration(float64(period) * float64(r.percent) / 100.0)

	for i := 0; i < r.cores; i++ {
		go func() {
			defer wg.Done()

			var burn uint64
			for {
				start := time.Now()
				for time.Since(start) < busy {
					burn = burn*1664525 + 1013904223 // tight loop to keep the core busy
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(period - busy):
				}
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()
	return ctx.Err()
}
