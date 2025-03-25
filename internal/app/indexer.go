package app

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/service/category"
	"github.com/stickpro/go-store/pkg/logger"
	utils "github.com/stickpro/go-store/pkg/util"
)

func initIndexer(ctx context.Context, service *service.Services, l logger.Logger) {
	if service.SearchService != nil {
		cat, err := service.CategoryService.GetCategoryWithPagination(ctx, category.GetDTO{
			Page:     utils.Pointer(uint32(1)),
			PageSize: utils.Pointer(uint32(100)),
		})
		if err != nil {
			l.Error("failed to get category", err)
		}
		fmt.Println(cat.Items)
		data, err := structToMap(cat.Items)

		err = service.SearchService.CreateIndex("category", data)
		if err != nil {
			l.Error("failed to create index", err)
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
