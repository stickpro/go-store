package repository_product_reviews

import (
	"context"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type ProductReviewWithPaginationParams struct {
	base.CommonFindParams
	ProductID uuid.NullUUID
}

type FindRow struct {
	*models.ProductReview
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params ProductReviewWithPaginationParams,
) (*base.FindResponseWithFullPagination[*models.ProductReview], error) {
	return base.Paginate[*models.ProductReview](ctx, s.db, params.CommonFindParams, base.PaginationConfig[*models.ProductReview]{
		TableName:    "product_reviews",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"title":      true,
			"created_at": true,
		},
	})
}

func (s *CustomQueries) GetByProductIDWithPaginate(
	ctx context.Context,
	params ProductReviewWithPaginationParams,
) (*base.FindResponseWithFullPagination[*models.ProductReview], error) {
	return base.Paginate[*models.ProductReview](ctx, s.db, params.CommonFindParams, base.PaginationConfig[*models.ProductReview]{
		TableName:    "product_reviews",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		WhereBuilder: func(sb *sqlbuilder.SelectBuilder) {
			sb.Where(
				sb.Equal("status", constant.ReviewApproved.String()),
				sb.Equal("product_id", params.ProductID),
			)
		},
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"title":      true,
			"created_at": true,
		},
	})
}
