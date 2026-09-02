package repository_product_variants

import (
	"context"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type VariantsWithPaginationParams struct {
	base.CommonFindParams
	// CategoryID filters variants that belong to this category or any of its
	// descendants, either by their primary category_id or via the
	// product_variant_categories junction.
	CategoryID *uuid.UUID
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
		WhereBuilder: func(sb *sqlbuilder.SelectBuilder) {
			if params.CategoryID != nil {
				sb.Where(
					"EXISTS (" +
						"SELECT 1 FROM category_paths cp " +
						"WHERE cp.ancestor_id = " + sb.Var(*params.CategoryID) + " " +
						"AND (cp.descendant_id = product_variants.category_id " +
						"OR EXISTS (" +
						"SELECT 1 FROM product_variant_categories pvc " +
						"WHERE pvc.product_variant_id = product_variants.id " +
						"AND pvc.category_id = cp.descendant_id)))",
				)
			}
		},
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
