package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stickpro/go-store/pkg/logger"
)

const (
	workerRestartBaseDelay = 1 * time.Second
	workerRestartMaxDelay  = 30 * time.Second
)

// worker is a supervised background loop: run blocks until ctx is cancelled,
// or returns early on failure. See asWorker for the common case of adapting a
// `func(context.Context)` (a *Worker.Run-style method) into one.
type worker struct {
	name string
	run  func(context.Context) error
}

// superviseWorkers starts every worker in its own goroutine and restarts it
// with exponential backoff if it panics or returns before ctx is cancelled.
// It returns a WaitGroup callers can wait on during shutdown.
func superviseWorkers(ctx context.Context, l logger.Logger, workers ...worker) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	for _, w := range workers {
		wg.Add(1)
		go func(w worker) {
			defer wg.Done()
			superviseWorker(ctx, l, w)
		}(w)
	}
	return wg
}

func superviseWorker(ctx context.Context, l logger.Logger, w worker) {
	delay := workerRestartBaseDelay
	for {
		if ctx.Err() != nil {
			return
		}

		start := time.Now()
		err := runWorkerOnce(ctx, w)
		switch {
		case ctx.Err() != nil:
			// Shutdown in progress - a returning worker is expected.
			return
		case err != nil:
			l.Errorw("background worker crashed, restarting", "worker", w.name, "uptime", time.Since(start).String(), "error", err)
		default:
			l.Warnw("background worker exited unexpectedly, restarting", "worker", w.name, "uptime", time.Since(start).String())
		}

		// A worker that ran for a while before dying gets a fresh backoff
		// instead of inheriting the previous failure's delay.
		if time.Since(start) >= workerRestartMaxDelay {
			delay = workerRestartBaseDelay
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		if delay *= 2; delay > workerRestartMaxDelay {
			delay = workerRestartMaxDelay
		}
	}
}

func runWorkerOnce(ctx context.Context, w worker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return w.run(ctx)
}

// asWorker adapts a run func that has no return value (it just blocks until
// ctx is cancelled) into a worker.
func asWorker(name string, run func(context.Context)) worker {
	return worker{name: name, run: func(ctx context.Context) error {
		run(ctx)
		return nil
	}}
}
