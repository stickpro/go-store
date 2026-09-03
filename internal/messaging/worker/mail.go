package worker

import (
	"context"
	"sync"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/service/mail"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

type MailWorker struct {
	queue       queue.IQueue
	mailSvc     mail.IMailService
	logger      logger.Logger
	concurrency int
}

func NewMailWorker(q queue.IQueue, mailSvc mail.IMailService, concurrency int, log logger.Logger) *MailWorker {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &MailWorker{queue: q, mailSvc: mailSvc, concurrency: concurrency, logger: log}
}

// Run starts concurrency worker goroutines and blocks until ctx is cancelled.
func (w *MailWorker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for range w.concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.loop(ctx)
		}()
	}
	wg.Wait()
	w.logger.Info("Mail workers stopped")
}

func (w *MailWorker) loop(ctx context.Context) {
	for {
		payload, err := w.queue.Pop(ctx, mail.QueueName)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Error("mail worker: queue pop error", "error", err)
			continue
		}
		if payload == nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		var task mail.Task
		if err := json.Unmarshal(payload, &task); err != nil {
			w.logger.Error("mail worker: unmarshal task", "error", err)
			continue
		}

		if err := w.mailSvc.Deliver(ctx, task); err != nil {
			if ctx.Err() != nil {
				return
			}
			w.mailSvc.HandleFailure(task, err)
		}
	}
}
