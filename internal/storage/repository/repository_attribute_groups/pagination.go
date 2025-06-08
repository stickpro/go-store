package repository_attribute_groups

import (
	"context"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params base.CommonFindParams,
) (*base.FindResponseWithFullPagination[*models.AttributeGroup], error) {
	return base.Paginate[*models.AttributeGroup](ctx, s.db, params, base.PaginationConfig[*models.AttributeGroup]{
		TableName:    "attribute",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
