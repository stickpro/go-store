package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

type IRelatedProduct interface {
	GetRelatedProducts(ctx context.Context, variantID uuid.UUID) ([]*models.ShortProduct, error)
	GetRelatedProductsBatch(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID][]*models.ShortProduct, error)
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
			Price:    p.PriceRetail,
			IsEnable: p.IsEnable,
			Image:    p.Image,
		})
	}
	return resp, nil
}

func (s *Service) GetRelatedProductsBatch(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID][]*models.ShortProduct, error) {
	rows, err := s.storage.Products().GetRelatedProductsByVariantIDs(ctx, variantIDs)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get related products batch", "error", parsedErr)
		return nil, parsedErr
	}

	result := make(map[uuid.UUID][]*models.ShortProduct, len(variantIDs))
	for _, row := range rows {
		result[row.VariantID] = append(result[row.VariantID], &models.ShortProduct{
			ID:       row.ID,
			Name:     row.Name,
			Slug:     row.Slug,
			Model:    row.Model,
			Price:    row.PriceRetail,
			IsEnable: row.IsEnable,
			Image:    row.Image,
		})
	}
	return result, nil
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
			Price:    p.PriceRetail,
			IsEnable: p.IsEnable,
			Image:    p.Image,
		})
	}
	return resp, nil
}

func (s *Service) SyncRelatedProducts(ctx context.Context, variantID uuid.UUID, relatedVariantIDs []uuid.UUID) error {
	err := s.storage.Products().SyncRelatedProducts(ctx, repository_products.SyncRelatedProductsParams{
		VariantID:         variantID,
		RelatedVariantIds: relatedVariantIDs,
	})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to sync related products", "error", parsedErr)
		return parsedErr
	}
	return nil
}
