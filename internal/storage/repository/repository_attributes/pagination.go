package repository_attributes

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/pkg/dbutils"
	"math"
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
	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder().
		Select("COUNT(1)").
		From("attributes")

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder().
		Select("*").
		From("attributes")

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(100))
	if err != nil {
		return nil, fmt.Errorf("pagination error: %w", err)
	}

	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "attributes.created_at"
	}
	if !params.IsAscOrdering {
		sb.Desc()
	}
	sb.OrderBy(orderBy)
	sb.Limit(int(limit)).Offset(int(offset)) // #nosec

	var items []*FindRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to fetch items: %w", err)
	}

	var totalCnt uint64
	pagingSQL, args := countSb.Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, pagingSQL, args...); err != nil {
		return nil, fmt.Errorf("failed to fetch total count: %w", err)
	}

	page := uint64(1)
	if params.Page != nil {
		page = uint64(*params.Page)
	}
	lastPage := uint64(1)
	if params.PageSize != nil {
		lastPage = uint64(math.Ceil(float64(totalCnt) / float64(*params.PageSize)))
	}

	return &base.FindResponseWithFullPagination[*FindRow]{
		Items: items,
		Pagination: base.FullPagingData{
			Total:    totalCnt,
			PageSize: uint64(limit),
			Page:     page,
			LastPage: lastPage,
		},
	}, nil
}
