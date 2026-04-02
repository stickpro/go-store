package worker

import (
	"context"
	"sync"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/messaging/tasks"
	"github.com/stickpro/go-store/internal/service/media"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

type ImageWorker struct {
	queue       queue.IQueue
	mediaSvc    media.IMediaService
	logger      logger.Logger
	concurrency int
}

func NewImageWorker(q queue.IQueue, mediaSvc media.IMediaService, concurrency int, log logger.Logger) *ImageWorker {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &ImageWorker{queue: q, mediaSvc: mediaSvc, concurrency: concurrency, logger: log}
}

// Run starts concurrency worker goroutines and blocks until ctx is cancelled.
func (w *ImageWorker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for range w.concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.loop(ctx)
		}()
	}
	wg.Wait()
	w.logger.Info("Image workers stopped")
}

func (w *ImageWorker) loop(ctx context.Context) {
	for {
		payload, err := w.queue.Pop(ctx, tasks.ImageSyncQueue)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Error("image worker: queue pop error", "error", err)
			continue
		}
		if payload == nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		var task tasks.ImageSyncTask
		if err := json.Unmarshal(payload, &task); err != nil {
			w.logger.Error("image worker: unmarshal task", "error", err)
			continue
		}

		if err := w.mediaSvc.SyncProductImages(ctx, task.ProductID, task.ImageMain, task.Images); err != nil {
			w.logger.Errorw("image worker: sync product images",
				"product_id", task.ProductID,
				"error", err,
			)
		}
	}
}
