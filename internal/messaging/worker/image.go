package worker

import (
	"context"
	"encoding/json"

	"github.com/stickpro/go-store/internal/messaging/tasks"
	"github.com/stickpro/go-store/internal/service/media"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

type ImageWorker struct {
	queue    queue.IQueue
	mediaSvc media.IMediaService
	logger   logger.Logger
}

func NewImageWorker(q queue.IQueue, mediaSvc media.IMediaService, log logger.Logger) *ImageWorker {
	return &ImageWorker{queue: q, mediaSvc: mediaSvc, logger: log}
}

func (w *ImageWorker) Run(ctx context.Context) {
	w.logger.Info("Image worker started")
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
