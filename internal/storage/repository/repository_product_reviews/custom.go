package repository_product_reviews

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
		params ProductReviewWithPaginationParams,
	) (*base.FindResponseWithFullPagination[*models.ProductReview], error)
	GetByProductIDWithPaginate(
		ctx context.Context,
		params ProductReviewWithPaginationParams,
	) (*base.FindResponseWithFullPagination[*models.ProductReview], error)
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
