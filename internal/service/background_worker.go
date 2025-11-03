package service

import (
	"log/slog"
	"sync"
)

// BackgroundWorker is a simple wrapper around a WaitGroup
type BackgroundWorker struct {
	wg     sync.WaitGroup
	logger *slog.Logger
}

// NewBackgroundWorker create new worker
func NewBackgroundWorker(logger *slog.Logger) *BackgroundWorker {
	return &BackgroundWorker{
		logger: logger,
	}
}

// Run launches a function in a new background goroutine.
// It uses a WaitGroup to track the number of active goroutines.
func (bw *BackgroundWorker) Run(fn func()) {
	bw.wg.Add(1)

	go func() {
		defer bw.wg.Done()

		// Recover from any panics in the background task
		defer func() {
			if err := recover(); err != nil {
				bw.logger.Error("recovered from panic in background worker", "error", err)
			}
		}()

		fn()
	}()
}

// Wait blocks until all background tasks have completed
func (bw *BackgroundWorker) Wait() {
	bw.wg.Wait()
}
