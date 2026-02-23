package repository_products

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type ProductsWithPaginationParams struct {
	base.CommonFindParams
	WithoutVariants bool
}

type FindRow struct {
	models.Product
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params ProductsWithPaginationParams,
) (*base.FindResponseWithFullPagination[*FindRow], error) {
	cfg := base.PaginationConfig[*FindRow]{
		TableName:    "products",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	}

	if params.WithoutVariants {
		cfg.WhereBuilder = func(sb *sqlbuilder.SelectBuilder) {
			sb.Where("NOT EXISTS (SELECT 1 FROM product_variants WHERE product_variants.product_id = products.id)")
		}
	}

	return base.Paginate[*FindRow](ctx, s.db, params.CommonFindParams, cfg)
}
