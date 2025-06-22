package repository_collections

import (
	"context"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type CollectionWithPaginationParams struct {
	base.CommonFindParams
}

type FindRow struct {
	*models.Collection
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params CollectionWithPaginationParams,
) (*base.FindResponseWithFullPagination[*models.Collection], error) {
	return base.Paginate[*models.Collection](ctx, s.db, params.CommonFindParams, base.PaginationConfig[*models.Collection]{
		TableName:    "collections",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
