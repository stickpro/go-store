package repository_attributes

import (
	"context"

	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params base.CommonFindParams,
) (*base.FindResponseWithFullPagination[*models.Attribute], error) {
	return base.Paginate[*models.Attribute](ctx, s.db, params, base.PaginationConfig[*models.Attribute]{
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
