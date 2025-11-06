package base

import (
	"context"
	"fmt"
	"math"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/stickpro/go-store/pkg/dbutils"
)

type PaginationConfig[R any] struct {
	TableName         string
	DefaultOrder      string
	MaxLimit          uint32
	AllowedFieldOrder map[string]bool
	WhereBuilder      func(sb *sqlbuilder.SelectBuilder)
}

func Paginate[R any](
	ctx context.Context,
	db pgxscan.Querier,
	params CommonFindParams,
	cfg PaginationConfig[R],
) (*FindResponseWithFullPagination[R], error) {
	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(cfg.MaxLimit))
	if err != nil {
		return nil, fmt.Errorf("pagination error: %w", err)
	}

	orderBy := params.OrderBy
	if !cfg.AllowedFieldOrder[orderBy] {
		orderBy = cfg.DefaultOrder
	}
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder().Select("*").From(cfg.TableName)
	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder().Select("COUNT(1)").From(cfg.TableName)

	if cfg.WhereBuilder != nil {
		cfg.WhereBuilder(sb)
		cfg.WhereBuilder(countSb)
	}

	if params.WithDeleted {
		sb.Where("deleted_at IS NOT NULL")
		countSb.Where("deleted_at IS NOT NULL")
	}

	if params.IsAscOrdering {
		sb.OrderBy(orderBy + " ASC")
	} else {
		sb.OrderBy(orderBy + " DESC")
	}

	sb.Limit(int(limit)).Offset(int(offset))

	sql, args := sb.Build()
	var items []R
	if err := pgxscan.Select(ctx, db, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("failed to fetch items: %w", err)
	}

	countSQL, countArgs := countSb.Build()
	var total uint64
	if err := pgxscan.Get(ctx, db, &total, countSQL, countArgs...); err != nil {
		return nil, fmt.Errorf("failed to fetch total count: %w", err)
	}

	page := uint64(1)
	if params.Page != nil {
		page = uint64(*params.Page)
	}
	lastPage := uint64(math.Ceil(float64(total) / float64(limit)))

	return &FindResponseWithFullPagination[R]{
		Items: items,
		Pagination: FullPagingData{
			Total:    total,
			PageSize: uint64(limit),
			Page:     page,
			LastPage: lastPage,
		},
	}, nil
}
