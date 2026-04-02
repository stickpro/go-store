package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/messaging/handler"
	"github.com/stickpro/go-store/internal/messaging/kafka"
	"github.com/stickpro/go-store/internal/messaging/worker"
	"github.com/stickpro/go-store/internal/server"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

func Run(ctx context.Context, conf *config.Config, l logger.Logger) {
	st, err := storage.InitStore(ctx, conf)
	if err != nil {
		l.Fatal("failed to init store", err)
		return
	}
	defer func() {
		if err := st.Close(); err != nil {
			l.Fatal("storage close error", err)
		}
	}()
	l.Info("Storage init")

	q, err := initQueue(conf)
	if err != nil {
		l.Fatal("failed to init queue", err)
		return
	}
	defer func() {
		if err := q.Close(); err != nil {
			l.Error("queue close error", err)
		}
	}()

	services, err := service.InitService(conf, l, st)
	if err != nil {
		l.Fatal("error start DI service", err)
	}
	defer func() {
		if err := services.Close(); err != nil {
			l.Fatal("services close error", err)
		}
	}()
	l.Info("Start DI service")

	srv := server.InitServer(conf, services, l)

	initIndexer(ctx, services, l, true)

	consumer, err := kafka.NewConsumer(conf.Kafka, l)
	if err != nil {
		l.Fatal("failed to init kafka consumer", err)
		return
	}
	defer consumer.Close()

	productHandler := handler.NewProductHandler(services.ProductService, services.AttributeService, q, l)

	go func() {
		l.Info("Start kafka consumer")
		if err := consumer.Run(ctx, productHandler.HandleProduct); err != nil {
			l.Error("kafka consumer stopped with error", err)
		}
	}()

	imageWorker := worker.NewImageWorker(q, services.MediaService, conf.Workers.ImageSync, l)
	go imageWorker.Run(ctx)

	serverErrCh := make(chan error, 1)
	go func() {
		defer close(serverErrCh)
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("error occurred while running http server", err)
			serverErrCh <- err
		}
	}()

	l.Info("Start http server")

	select {
	case <-ctx.Done():
		l.Info("Shutdown signal received")
		if err := srv.Stop(); err != nil {
			l.Error("failed to stop server", err)
		}
	case err := <-serverErrCh:
		l.Error("Server crashed, shutting down", err)
	}
}

func initQueue(conf *config.Config) (queue.IQueue, error) {
	switch conf.KeyValue.Engine {
	case config.KeyValueEngineRedis:
		return queue.NewRedisQueue(conf.Redis.URL())
	case config.KeyValueEngineInMemory:
		return queue.NewInMemoryQueue(), nil
	default:
		return nil, fmt.Errorf("queue: unsupported engine %q", conf.KeyValue.Engine)
	}
}
