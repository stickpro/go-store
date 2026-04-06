package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_variant_categories"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

type IVariantCategoryService interface {
	GetVariantCategories(ctx context.Context, variantID uuid.UUID) ([]*dto.VariantCategoryDTO, error)
	SyncVariantCategories(ctx context.Context, d dto.SyncVariantCategoriesDTO) error
}

func (s *Service) GetVariantCategories(ctx context.Context, variantID uuid.UUID) ([]*dto.VariantCategoryDTO, error) {
	rows, err := s.storage.ProductVariantCategories().GetByVariantID(ctx, variantID)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}

	result := make([]*dto.VariantCategoryDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, &dto.VariantCategoryDTO{
			CategoryID:       row.CategoryID,
			CategoryName:     row.CategoryName,
			CategorySlug:     row.CategorySlug,
			CategoryIsEnable: row.CategoryIsEnable,
		})
	}
	return result, nil
}

func (s *Service) SyncVariantCategories(ctx context.Context, d dto.SyncVariantCategoriesDTO) error { //nolint:dupl
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.storage.ProductVariantCategories(repository.WithTx(tx)).RemoveAll(ctx, d.VariantID)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to remove variant categories", parsedErr)
			return parsedErr
		}

		if len(d.CategoryIDs) == 0 {
			return nil
		}

		err = s.storage.ProductVariantCategories(repository.WithTx(tx)).AddBatch(ctx, repository_product_variant_categories.AddBatchParams{
			ProductVariantID: d.VariantID,
			CategoryIds:      d.CategoryIDs,
		})
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to add variant categories", parsedErr)
			return parsedErr
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
