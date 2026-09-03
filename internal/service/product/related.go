package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

type IRelatedProduct interface {
	GetRelatedProducts(ctx context.Context, variantID uuid.UUID) ([]*dto.VariantCardDTO, error)
	GetRelatedProductsBatch(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID][]*dto.VariantCardDTO, error)
	GetRelatedProductsBySlug(ctx context.Context, slug string) ([]*dto.VariantCardDTO, error)
	SyncRelatedProducts(ctx context.Context, variantID uuid.UUID, relatedVariantIDs []uuid.UUID) error
}

func (s *Service) GetRelatedProducts(ctx context.Context, variantID uuid.UUID) ([]*dto.VariantCardDTO, error) {
	products, err := s.storage.Products().GetRelatedProductsByVariantID(ctx, variantID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get related products", err)
		return nil, parsedErr
	}
	resp := make([]*dto.VariantCardDTO, 0, len(products))
	for _, p := range products {
		resp = append(resp, &dto.VariantCardDTO{
			ID:             p.ID,
			ProductID:      p.ProductID,
			Name:           p.Name,
			Slug:           p.Slug,
			Model:          p.Model,
			PriceRetail:    p.PriceRetail,
			PriceBusiness:  p.PriceBusiness,
			PriceWholesale: p.PriceWholesale,
			StockStatus:    p.StockStatus,
			IsEnable:       p.IsEnable,
			Image:          s.shortImage(p.ImageID, p.ImagePath, p.ImageWidth, p.ImageHeight, p.Name),
		})
	}
	return resp, nil
}

func (s *Service) GetRelatedProductsBatch(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID][]*dto.VariantCardDTO, error) {
	rows, err := s.storage.Products().GetRelatedProductsByVariantIDs(ctx, variantIDs)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get related products batch", "error", parsedErr)
		return nil, parsedErr
	}

	result := make(map[uuid.UUID][]*dto.VariantCardDTO, len(variantIDs))
	for _, row := range rows {
		result[row.VariantID] = append(result[row.VariantID], &dto.VariantCardDTO{
			ID:             row.ID,
			ProductID:      row.ProductID,
			Name:           row.Name,
			Slug:           row.Slug,
			Model:          row.Model,
			PriceRetail:    row.PriceRetail,
			PriceBusiness:  row.PriceBusiness,
			PriceWholesale: row.PriceWholesale,
			StockStatus:    row.StockStatus,
			IsEnable:       row.IsEnable,
			Image:          s.shortImage(row.ImageID, row.ImagePath, row.ImageWidth, row.ImageHeight, row.Name),
		})
	}
	return result, nil
}

func (s *Service) GetRelatedProductsBySlug(ctx context.Context, slug string) ([]*dto.VariantCardDTO, error) {
	products, err := s.storage.Products().GetRelatedProductsBySlug(ctx, slug)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get related products by slug", err)
		return nil, parsedErr
	}
	resp := make([]*dto.VariantCardDTO, 0, len(products))
	for _, p := range products {
		resp = append(resp, &dto.VariantCardDTO{
			ID:             p.ID,
			ProductID:      p.ProductID,
			Name:           p.Name,
			Slug:           p.Slug,
			Model:          p.Model,
			PriceRetail:    p.PriceRetail,
			PriceBusiness:  p.PriceBusiness,
			PriceWholesale: p.PriceWholesale,
			StockStatus:    p.StockStatus,
			IsEnable:       p.IsEnable,
			Image:          s.shortImage(p.ImageID, p.ImagePath, p.ImageWidth, p.ImageHeight, p.Name),
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

// shortImage builds the ImageDTO for a related/listing row, or nil when the product has no image.
func (s *Service) shortImage(id uuid.NullUUID, imgPath pgtype.Text, w, h pgtype.Int4, alt string) *models.ImageDTO {
	if !id.Valid || !imgPath.Valid {
		return nil
	}
	img := dto.NewImageDTO(id.UUID, imgPath.String, w.Int32, h.Int32, alt, s.cfg.Images.ResolvedPresets())
	return &img
}
