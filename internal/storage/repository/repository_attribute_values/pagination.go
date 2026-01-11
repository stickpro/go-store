package repository_attribute_values

import (
	"context"

	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params base.CommonFindParams,
) (*base.FindResponseWithFullPagination[*models.AttributeValue], error) {
	return base.Paginate[*models.AttributeValue](ctx, s.db, params, base.PaginationConfig[*models.AttributeValue]{
		TableName:    "attribute_values",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
