package app

import (
	"context"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/messaging/handler"
	"github.com/stickpro/go-store/internal/messaging/kafka"
	queueworker "github.com/stickpro/go-store/internal/messaging/worker"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

// buildWorkers assembles every background loop the app runs alongside the
// HTTP server: the Kafka product consumer, the queue-backed image/mail
// workers, and the CDEK delivery points cache refresher. Each is supervised
// independently (see worker.go) so a crash in one doesn't take down the rest,
// and it gets restarted with backoff instead of killing the process.
func buildWorkers(
	services *service.Services,
	conf *config.Config,
	q queue.IQueue,
	consumer *kafka.Consumer,
	productHandler *handler.ProductHandler,
	l logger.Logger,
) []worker {
	var workers []worker
	add := func(name string, enabled bool, run func(context.Context)) {
		if !enabled {
			return
		}
		workers = append(workers, asWorker(name, run))
	}

	add("kafka_consumer", consumer != nil, func(ctx context.Context) {
		if err := consumer.Run(ctx, productHandler.HandleProduct); err != nil {
			l.Error("kafka consumer stopped with error", err)
		}
	})

	add("image_worker", services.MediaService != nil, func(ctx context.Context) {
		queueworker.NewImageWorker(q, services.MediaService, conf.Workers.ImageSync, l).Run(ctx)
	})

	add("mail_worker", services.MailService != nil, func(ctx context.Context) {
		queueworker.NewMailWorker(q, services.MailService, conf.Workers.MailSend, l).Run(ctx)
	})

	add("cdek_delivery_points_cache", services.CDEKService != nil, func(ctx context.Context) {
		services.CDEKService.RunCacheRefresher(ctx)
	})

	return workers
}
