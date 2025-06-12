package app

import (
	"context"
	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/pkg/logger"
)

func initIndexer(ctx context.Context, service *service.Services, l logger.Logger, reindex ...bool) {
	forceReindex := false
	if len(reindex) > 0 {
		forceReindex = reindex[0]
	}

	if service.GeoService != nil {
		shouldCreate := forceReindex
		if !forceReindex {
			exists, err := service.SearchService.CheckIndex(constant.CitiesIndex)
			if err != nil {
				l.Error("failed to check cities index", err)
			}
			shouldCreate = !exists
		}

		if shouldCreate {
			cities, err := service.GeoService.GetAllCity(ctx)
			if err != nil {
				l.Error("failed to get cities", err)
			} else {
				data, err := structToMap(cities)
				if err != nil {
					l.Error("failed to convert cities", err)
				} else {
					opts := searchtypes.IndexOptions{
						SearchableAttributes: []string{"city", "address", "region"},
					}
					err = service.SearchService.CreateIndex(constant.CitiesIndex, data, opts)
					if err != nil {
						l.Error("failed to create city index", err)
					}
				}
			}
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
