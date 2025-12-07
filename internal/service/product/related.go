package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

type IRelatedProduct interface {
	GetRelatedProduct(ctx context.Context, productID uuid.UUID) ([]*models.ShortProduct, error)
	SyncRelatedProduct(ctx context.Context, productID uuid.UUID, relatedProductIDs []uuid.UUID) error
}

func (s *Service) GetRelatedProduct(ctx context.Context, productID uuid.UUID) ([]*models.ShortProduct, error) {
	products, err := s.storage.Products().GetRelatedProductsByProductID(ctx, productID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get product media", err)
		return nil, parsedErr
	}
	resp := make([]*models.ShortProduct, 0, len(products))
	for _, p := range products {
		resp = append(resp, &models.ShortProduct{
			ID:       p.ID,
			Name:     p.Name,
			Slug:     p.Slug,
			Model:    p.Model,
			Price:    p.Price,
			IsEnable: p.IsEnable,
			Image:    p.Image,
		})
	}
	return resp, nil
}

func (s *Service) SyncRelatedProduct(ctx context.Context, productID uuid.UUID, relatedProductIDs []uuid.UUID) error {
	return repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.storage.Products(repository.WithTx(tx)).DeleteRelatedProducts(ctx, productID)
		if err != nil {
			return err
		}
		if len(relatedProductIDs) == 0 {
			return nil
		}
		err = s.storage.Products(repository.WithTx(tx)).AddRelatedProducts(ctx, repository_products.AddRelatedProductsParams{
			ProductID:         productID,
			RelatedProductIds: relatedProductIDs,
		})
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to update related product", err)
			return parsedErr
		}
		return nil
	})
}
