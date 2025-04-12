package repository_attribute_groups

import (
	"context"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type AttributeGroupsWithPaginationParams struct {
	base.CommonFindParams
}

type FindRow struct {
	*models.AttributeGroup
}

func (s *CustomQueries) GetWithPaginate(
	ctx context.Context,
	params base.CommonFindParams,
) (*base.FindResponseWithFullPagination[*FindRow], error) {
	return base.Paginate[models.AttributeGroup, *FindRow](ctx, s.db, params, base.PaginationConfig[models.AttributeGroup, *FindRow]{
		TableName:    "attribute_groups",
		DefaultOrder: "created_at",
		MaxLimit:     100,
		AllowedFieldOrder: map[string]bool{
			"id":         true,
			"name":       true,
			"created_at": true,
		},
	})
}
