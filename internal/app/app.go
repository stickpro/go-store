package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/messaging/handler"
	"github.com/stickpro/go-store/internal/messaging/kafka"
	"github.com/stickpro/go-store/internal/server"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/imageprocessor"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

// workerShutdownTimeout bounds how long Run waits for background workers to
// stop after a shutdown signal, before returning anyway.
const workerShutdownTimeout = 25 * time.Second

func Run(ctx context.Context, conf *config.Config, l logger.Logger) {
	imageprocessor.Startup()
	defer imageprocessor.Shutdown()

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

	services, err := service.InitService(conf, l, st, q)
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

	workerCtx, stopWorkers := context.WithCancel(ctx)
	defer stopWorkers()
	workersWG := superviseWorkers(workerCtx, l, buildWorkers(services, conf, q, consumer, productHandler, l)...)

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

	stopWorkers()
	waitWorkers(workersWG, workerShutdownTimeout, l)
}

// waitWorkers blocks until wg is done or timeout elapses, whichever comes first.
func waitWorkers(wg *sync.WaitGroup, timeout time.Duration, l logger.Logger) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		l.Info("background workers stopped")
	case <-time.After(timeout):
		l.Warnw("background workers did not stop before shutdown timeout", "timeout", timeout.String())
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
