package app

import (
	"context"

	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/pkg/logger"
)

func initIndexer(ctx context.Context, service *service.Services, l logger.Logger, reindex ...bool) {
	forceReindex := true
	if len(reindex) > 0 {
		forceReindex = reindex[0]
	}
	if service.GeoService != nil {
		err := service.GeoService.CreateCityIndex(ctx, forceReindex)
		if err != nil {
			l.Error("failed to create city index", "error", err)
		}
	}

	if service.ProductService != nil {
		err := service.ProductService.CreateProductIndex(ctx, forceReindex)
		if err != nil {
			l.Error("failed to create product index", "error", err)
		}
	}
}
