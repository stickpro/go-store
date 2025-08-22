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
		if storageCloseErr := st.Close(); storageCloseErr != nil {
			l.Fatal("storage close error", storageCloseErr)
		}
	}()

	l.Info("Storage init")

	services, err := service.InitService(conf, l, st)
	if err != nil {
		l.Fatal("error start DI service", err)
	}
	defer func() {
		if serviceCloseErr := services.Close(); serviceCloseErr != nil {
			l.Fatal("services close error", serviceCloseErr)
		}
	}()
	l.Info("Start DI service")

	srv := server.InitServer(conf, services, l)

	initIndexer(ctx, services, l)
	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			l.Error("error occurred while running http server", err)
		}
	}()

	l.Info("Start http server")

	if err := srv.Stop(); err != nil {
		l.Error("failed to stop server", err)
	}

	<-ctx.Done()
}
