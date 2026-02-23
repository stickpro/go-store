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
	GetRelatedProducts(ctx context.Context, variantID uuid.UUID) ([]*models.ShortProduct, error)
	GetRelatedProductsBySlug(ctx context.Context, slug string) ([]*models.ShortProduct, error)
	SyncRelatedProducts(ctx context.Context, variantID uuid.UUID, relatedVariantIDs []uuid.UUID) error
}

func (s *Service) GetRelatedProducts(ctx context.Context, variantID uuid.UUID) ([]*models.ShortProduct, error) {
	products, err := s.storage.Products().GetRelatedProductsByVariantID(ctx, variantID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get related products", err)
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

func (s *Service) GetRelatedProductsBySlug(ctx context.Context, slug string) ([]*models.ShortProduct, error) {
	products, err := s.storage.Products().GetRelatedProductsBySlug(ctx, slug)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get related products by slug", err)
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

func (s *Service) SyncRelatedProducts(ctx context.Context, variantID uuid.UUID, relatedVariantIDs []uuid.UUID) error {
	return repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.storage.Products(repository.WithTx(tx)).DeleteRelatedProducts(ctx, variantID)
		if err != nil {
			return err
		}
		if len(relatedVariantIDs) == 0 {
			return nil
		}
		err = s.storage.Products(repository.WithTx(tx)).AddRelatedProducts(ctx, repository_products.AddRelatedProductsParams{
			VariantID:         variantID,
			RelatedVariantIds: relatedVariantIDs,
		})
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to sync related products", err)
			return parsedErr
		}
		return nil
	})
}
