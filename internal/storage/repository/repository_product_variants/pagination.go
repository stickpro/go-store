package repository_product_variants

import (
	"context"

	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type VariantsWithPaginationParams struct {
	base.CommonFindParams
}

type FindRow struct {
	models.ProductVariant
} //	@name	ProductVariantListItem

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params VariantsWithPaginationParams,
) (*base.FindResponseWithFullPagination[*FindRow], error) {
	return base.Paginate[*FindRow](ctx, s.db, params.CommonFindParams, base.PaginationConfig[*FindRow]{
		TableName:    "product_variants",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
