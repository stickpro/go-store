package app

import (
	"context"
	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/pkg/logger"
)

func initIndexer(ctx context.Context, service *service.Services, l logger.Logger, reindex ...bool) {
	forceReindex := false
	if len(reindex) > 0 {
		forceReindex = reindex[0]
	}
	if service.GeoService != nil {
		err := service.GeoService.CreateCityIndex(ctx, forceReindex)
		if err != nil {
			l.Error("failed to create city index", "error", err)
		}
	}
}

func structToMap(data any) ([]map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
