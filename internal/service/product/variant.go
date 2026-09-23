package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/dto/mapper"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_variants"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type IVariantProductService interface { //nolint:interfacebloat
	// Variant CRUD
	CreateProductVariant(ctx context.Context, productID uuid.UUID, d dto.CreateProductVariantDTO) (*models.ProductVariant, error)
	GetProductVariantByID(ctx context.Context, variantID uuid.UUID) (*models.ProductVariant, error)
	GetProductVariants(ctx context.Context, productID uuid.UUID) ([]*models.ProductVariant, error)
	UpdateProductVariant(ctx context.Context, variantID uuid.UUID, d dto.UpdateProductVariantDTO) (*models.ProductVariant, error)
	DeleteProductVariant(ctx context.Context, variantID uuid.UUID) error
	GetVariantsWithPagination(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*repository_product_variants.FindRow], error)
	GetEnrichedVariantsWithPagination(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*dto.EnrichedVariantDTO], error)
	// GetEnrichedVariantsByCategoryWithPagination reads straight from the DB (always consistent) — used by admin.
	GetEnrichedVariantsByCategoryWithPagination(ctx context.Context, categoryID uuid.UUID, d dto.GetDTO) (*base.FindResponseWithFullPagination[*dto.EnrichedVariantDTO], error)
	// SearchVariantsByCategory reads from the search index (pre-enriched docs, faceted filtering) — used by the storefront.
	SearchVariantsByCategory(ctx context.Context, d dto.CategoryProductsFilterDTO) (*dto.VariantListDTO, error)
	// SearchVariants is the full-text storefront search over the variant index with optional
	// category / price / manufacturer / attribute filtering.
	SearchVariants(ctx context.Context, d dto.VariantSearchDTO) (*dto.VariantListDTO, error)
	// GetCategoryFilters returns the full filter set (with counts) for a category listing — used by the storefront.
	GetCategoryFilters(ctx context.Context, categorySlug string) (*dto.CategoryFiltersDTO, error)

	// Slug-based methods
	GetVariantBySlug(ctx context.Context, slug string) (*models.ProductVariant, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetProductWithMediaByVariantSlug(ctx context.Context, slug string) (*dto.ProductWithMediaDTO, error)
	GetProductAttributesBySlug(ctx context.Context, slug string) ([]*dto.AttributeGroupWithValuesDTO, error)
}

func (s *Service) CreateProductVariant(ctx context.Context, productID uuid.UUID, d dto.CreateProductVariantDTO) (*models.ProductVariant, error) {
	if _, err := s.storage.Products().GetByID(ctx, productID); err != nil {
		return nil, pgerror.ParseError(err)
	}

	params := repository_product_variants.CreateParams{
		ProductID:       productID,
		CategoryID:      d.CategoryID,
		Name:            d.Name,
		Slug:            d.Slug,
		Description:     pgtypeutils.EncodeText(d.Description),
		MetaTitle:       pgtypeutils.EncodeText(d.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(d.MetaH1),
		MetaDescription: pgtypeutils.EncodeText(d.MetaDescription),
		MetaKeyword:     pgtypeutils.EncodeText(d.MetaKeyword),
		SortOrder:       d.SortOrder,
		IsEnable:        d.IsEnable,
		Viewed:          0,
	}
	variant, err := s.storage.ProductVariants().Create(ctx, params)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	return variant, nil
}

func (s *Service) GetProductVariantByID(ctx context.Context, variantID uuid.UUID) (*models.ProductVariant, error) {
	variant, err := s.storage.ProductVariants().Get(ctx, variantID)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	return variant, nil
}

func (s *Service) GetProductVariants(ctx context.Context, productID uuid.UUID) ([]*models.ProductVariant, error) {
	variants, err := s.storage.ProductVariants().GetByProductID(ctx, productID)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	return variants, nil
}

func (s *Service) UpdateProductVariant(ctx context.Context, variantID uuid.UUID, d dto.UpdateProductVariantDTO) (*models.ProductVariant, error) {
	current, err := s.storage.ProductVariants().Get(ctx, variantID)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}

	params := repository_product_variants.UpdateParams{
		ID:              variantID,
		ProductID:       current.ProductID,
		CategoryID:      d.CategoryID,
		Name:            d.Name,
		Slug:            d.Slug,
		Description:     pgtypeutils.EncodeText(d.Description),
		MetaTitle:       pgtypeutils.EncodeText(d.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(d.MetaH1),
		MetaDescription: pgtypeutils.EncodeText(d.MetaDescription),
		MetaKeyword:     pgtypeutils.EncodeText(d.MetaKeyword),
		SortOrder:       d.SortOrder,
		IsEnable:        d.IsEnable,
		Viewed:          current.Viewed,
	}
	variant, err := s.storage.ProductVariants().Update(ctx, params)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	return variant, nil
}

func (s *Service) DeleteProductVariant(ctx context.Context, variantID uuid.UUID) error {
	if err := s.storage.ProductVariants().Delete(ctx, variantID); err != nil {
		return pgerror.ParseError(err)
	}
	return nil
}

func (s *Service) GetVariantsWithPagination(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*repository_product_variants.FindRow], error) {
	commonParams := base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}

	variantRows, err := s.storage.ProductVariants().GetWithPaginate(ctx, repository_product_variants.VariantsWithPaginationParams{
		CommonFindParams: *commonParams,
	})
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	return variantRows, nil
}

func (s *Service) GetEnrichedVariantsWithPagination(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*dto.EnrichedVariantDTO], error) {
	variantRows, err := s.GetVariantsWithPagination(ctx, d)
	if err != nil {
		return nil, err
	}

	return s.enrichVariantRows(ctx, variantRows), nil
}

// GetEnrichedVariantsByCategoryWithPagination returns a paginated list of enriched variants
// that belong to the category or any of its descendants, read directly from the database.
func (s *Service) GetEnrichedVariantsByCategoryWithPagination(
	ctx context.Context,
	categoryID uuid.UUID,
	d dto.GetDTO,
) (*base.FindResponseWithFullPagination[*dto.EnrichedVariantDTO], error) {
	commonParams := base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}

	variantRows, err := s.storage.ProductVariants().GetWithPaginate(ctx, repository_product_variants.VariantsWithPaginationParams{
		CommonFindParams: *commonParams,
		CategoryID:       &categoryID,
	})
	if err != nil {
		return nil, pgerror.ParseError(err)
	}

	return s.enrichVariantRows(ctx, variantRows), nil
}

// enrichVariantRows augments paginated variant rows with product-level fields (prices, stock, manufacturer).
func (s *Service) enrichVariantRows(
	ctx context.Context,
	variantRows *base.FindResponseWithFullPagination[*repository_product_variants.FindRow],
) *base.FindResponseWithFullPagination[*dto.EnrichedVariantDTO] {
	productCache := make(map[uuid.UUID]*models.Product)
	items := make([]*dto.EnrichedVariantDTO, 0, len(variantRows.Items))

	for _, row := range variantRows.Items {
		product, ok := productCache[row.ProductID]
		if !ok {
			var err error
			product, err = s.storage.Products().GetByID(ctx, row.ProductID)
			if err != nil {
				s.logger.Warn("Failed to get product for variant", "variant_id", row.ID, "error", err)
				continue
			}
			productCache[row.ProductID] = product
		}

		items = append(items, &dto.EnrichedVariantDTO{
			ProductVariant: &row.ProductVariant,
			PriceRetail:    product.PriceRetail,
			PriceBusiness:  product.PriceBusiness,
			PriceWholesale: product.PriceWholesale,
			ManufacturerID: product.ManufacturerID,
			StockStatus:    product.StockStatus,
		})
	}

	return &base.FindResponseWithFullPagination[*dto.EnrichedVariantDTO]{
		Items:      items,
		Pagination: variantRows.Pagination,
	}
}

func (s *Service) GetVariantBySlug(ctx context.Context, slug string) (*models.ProductVariant, error) {
	variant, err := s.storage.ProductVariants().GetBySlug(ctx, slug)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	return variant, nil
}

func (s *Service) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	prd, err := s.storage.Products().GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return prd, nil
}

func (s *Service) GetProductWithMediaByVariantSlug(ctx context.Context, slug string) (*dto.ProductWithMediaDTO, error) {
	variant, err := s.storage.ProductVariants().GetBySlug(ctx, slug)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	prd, err := s.storage.Products().GetByID(ctx, variant.ProductID)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	media, err := s.storage.Products().GetMediaByProductID(ctx, prd.ID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get product media", err)
		return nil, parsedErr
	}

	return &dto.ProductWithMediaDTO{
		Product: prd,
		Variant: variant,
		Medium:  media,
	}, nil
}

func (s *Service) GetProductAttributesBySlug(ctx context.Context, slug string) ([]*dto.AttributeGroupWithValuesDTO, error) {
	prd, err := s.GetProductBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	rawAttributes, err := s.storage.ProductAttributeValues().GetByProductID(ctx, prd.ID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get product attributes", parsedErr)
		return nil, parsedErr
	}

	return mapper.MapProductAttributesToGroupedDTO(rawAttributes), nil
}
