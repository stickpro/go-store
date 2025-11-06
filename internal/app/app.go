package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/server"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
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

	initIndexer(ctx, services, l)

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
