package repository_categories

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

type ICustomQueries interface {
	Querier
	GetWithPaginate(
		ctx context.Context,
		params CategoryWithPaginationParams,
	) (*base.FindResponseWithFullPagination[*models.Category], error)
}

type CustomQueries struct {
	*Queries
	psql DBTX
}

func NewCustom(psql DBTX) *CustomQueries {
	return &CustomQueries{
		Queries: New(psql),
		psql:    psql,
	}
}

func (s *CustomQueries) WithTx(tx pgx.Tx) *CustomQueries {
	return &CustomQueries{
		Queries: New(tx),
		psql:    tx,
	}
}
