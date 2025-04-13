package repository_products

import (
	"context"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type ProductsWithPaginationParams struct {
	base.CommonFindParams
}

type FindRow struct {
	models.Product
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params ProductsWithPaginationParams,
) (*base.FindResponseWithFullPagination[*FindRow], error) {
	return base.Paginate[models.Product, *FindRow](ctx, s.db, params.CommonFindParams, base.PaginationConfig[models.Product, *FindRow]{
		TableName:    "products",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
