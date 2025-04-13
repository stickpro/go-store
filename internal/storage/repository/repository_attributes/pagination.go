package repository_attributes

import (
	"context"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type AttributesWithPaginationParams struct {
	base.CommonFindParams
}

type FindRow struct {
	models.Attribute
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params AttributesWithPaginationParams,
) (*base.FindResponseWithFullPagination[*FindRow], error) {
	return base.Paginate[models.AttributeGroup, *FindRow](ctx, s.db, params.CommonFindParams, base.PaginationConfig[models.AttributeGroup, *FindRow]{
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
