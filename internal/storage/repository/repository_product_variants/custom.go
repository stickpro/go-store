package repository_product_variants

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/storage/base"
)

type ICustomQueries interface {
	Querier
	GetWithPaginate(
		ctx context.Context,
		params VariantsWithPaginationParams,
	) (*base.FindResponseWithFullPagination[*FindRow], error)
}

type CustomQueries struct {
	*Queries
}

func NewCustom(psql DBTX) *CustomQueries {
	return &CustomQueries{
		Queries: New(psql),
	}
}

func (s *CustomQueries) WithTx(tx pgx.Tx) *CustomQueries {
	return &CustomQueries{
		Queries: New(tx),
	}
}
