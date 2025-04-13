package repository_manufacturers

import (
	"context"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type ManufacturersWithPaginationParams struct {
	base.CommonFindParams
}

type FindRow struct {
	models.Manufacturer
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params ManufacturersWithPaginationParams,
) (*base.FindResponseWithFullPagination[*FindRow], error) {
	return base.Paginate[models.Manufacturer, *FindRow](ctx, s.db, params.CommonFindParams, base.PaginationConfig[models.Manufacturer, *FindRow]{
		TableName:    "manufactures",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
