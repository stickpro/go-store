package repository_attributes

import (
	"context"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type AttributesWithPaginationParams struct {
	base.CommonFindParams
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params AttributesWithPaginationParams,
) (*base.FindResponseWithFullPagination[*models.Attribute], error) {
	return base.Paginate[*models.Attribute](ctx, s.db, params.CommonFindParams, base.PaginationConfig[*models.Attribute]{
		TableName:    "attributes",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
