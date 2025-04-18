package repository_categories

import (
	"context"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type CategoryWithPaginationParams struct {
	base.CommonFindParams
}

type FindRow struct {
	*models.Category
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params CategoryWithPaginationParams,
) (*base.FindResponseWithFullPagination[*FindRow], error) {
	return base.Paginate[models.Category, *FindRow](ctx, s.db, params.CommonFindParams, base.PaginationConfig[models.Category, *FindRow]{
		TableName:    "categories",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
