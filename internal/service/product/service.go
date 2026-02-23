package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/dto/mapper"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_attribute_values"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_variants"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type IProductService interface {
	// Base product CRUD
	CreateProduct(ctx context.Context, d dto.CreateProductDTO) (*models.Product, *models.ProductVariant, error)
	UpdateProduct(ctx context.Context, d dto.UpdateProductDTO) (*models.Product, *models.ProductVariant, error)
	// UpsertProductByExternalID
	UpsertProductByExternalID(ctx context.Context, externalID string, d dto.ProductUpsertDTO) (*models.Product, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetProductByExternalID(ctx context.Context, externalID string) (*models.Product, error)
	GetProductWithMediaByID(ctx context.Context, id uuid.UUID) (*dto.ProductWithMediaDTO, error)
	GetProductWithPagination(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*repository_products.FindRow], error)
	GetProductsWithoutVariants(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*repository_products.FindRow], error)

	// Attributes (by product ID)
	GetProductAttributesByID(ctx context.Context, id uuid.UUID) ([]*dto.AttributeGroupWithValuesDTO, error)
	SyncProductAttributes(ctx context.Context, d dto.SyncAttributeProductDTO) error

	// Indexing
	CreateProductIndex(ctx context.Context, reindex bool) error

	// Related products
	IRelatedProduct

	// Variant operations (slug-based lookups + variant CRUD)
	IVariantProductService
}

type Service struct {
	cfg           *config.Config
	logger        logger.Logger
	storage       storage.IStorage
	searchService searchtypes.ISearchService
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage, ss searchtypes.ISearchService) *Service {
	return &Service{
		cfg:           cfg,
		logger:        logger,
		storage:       storage,
		searchService: ss,
	}
}

func (s *Service) CreateProduct(ctx context.Context, d dto.CreateProductDTO) (*models.Product, *models.ProductVariant, error) {
	productParams := repository_products.CreateParams{
		ManufacturerID: d.ManufacturerID,
		Model:          d.Model,
		Sku:            pgtypeutils.EncodeText(d.Sku),
		Upc:            pgtypeutils.EncodeText(d.Upc),
		Ean:            pgtypeutils.EncodeText(d.Ean),
		Jan:            pgtypeutils.EncodeText(d.Jan),
		Isbn:           pgtypeutils.EncodeText(d.Isbn),
		Mpn:            pgtypeutils.EncodeText(d.Mpn),
		Location:       pgtypeutils.EncodeText(d.Location),
		Quantity:       d.Quantity,
		StockStatus:    d.StockStatus,
		Price:          d.Price,
		Weight:         d.Weight,
		Length:         d.Length,
		Width:          d.Width,
		Height:         d.Height,
		Subtract:       d.Subtract,
		Minimum:        d.Minimum,
		SortOrder:      d.SortOrder,
		IsEnable:       d.IsEnable,
	}

	var prd *models.Product
	var variant *models.ProductVariant
	var err error

	err = repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		prd, err = s.storage.Products(repository.WithTx(tx)).Create(ctx, productParams)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to create product", parsedErr)
			return parsedErr
		}

		variantParams := repository_product_variants.CreateParams{
			ProductID:       prd.ID,
			CategoryID:      d.Variant.CategoryID,
			Name:            d.Variant.Name,
			Slug:            d.Variant.Slug,
			Description:     pgtypeutils.EncodeText(d.Variant.Description),
			MetaTitle:       pgtypeutils.EncodeText(d.Variant.MetaTitle),
			MetaH1:          pgtypeutils.EncodeText(d.Variant.MetaH1),
			MetaDescription: pgtypeutils.EncodeText(d.Variant.MetaDescription),
			MetaKeyword:     pgtypeutils.EncodeText(d.Variant.MetaKeyword),
			Image:           pgtypeutils.EncodeText(d.Variant.Image),
			SortOrder:       d.Variant.SortOrder,
			IsEnable:        d.Variant.IsEnable,
			Viewed:          0,
		}
		variant, err = s.storage.ProductVariants(repository.WithTx(tx)).Create(ctx, variantParams)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to create product variant", parsedErr)
			return parsedErr
		}

		for _, mediaID := range d.MediaIDs {
			err := s.storage.Products(repository.WithTx(tx)).CreateProductMedia(ctx, repository_products.CreateProductMediaParams{
				ProductID: prd.ID,
				MediaID:   *mediaID,
			})
			if err != nil {
				parsedErr := pgerror.ParseError(err)
				s.logger.Error("failed to create product media", parsedErr)
				return parsedErr
			}
		}

		err = s.IndexProduct(ctx, prd)
		if err != nil {
			s.logger.Error("failed to index product", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return prd, variant, nil
}

func (s *Service) GetProductWithPagination(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*repository_products.FindRow], error) {
	commonParams := base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}

	prds, err := s.storage.Products().GetWithPaginate(ctx, repository_products.ProductsWithPaginationParams{
		CommonFindParams: *commonParams,
	})
	if err != nil {
		return nil, err
	}
	return prds, nil
}

func (s *Service) GetProductsWithoutVariants(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*repository_products.FindRow], error) {
	commonParams := base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}

	return s.storage.Products().GetWithPaginate(ctx, repository_products.ProductsWithPaginationParams{
		CommonFindParams: *commonParams,
		WithoutVariants:  true,
	})
}

func (s *Service) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	prd, err := s.storage.Products().GetByID(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Debugf("failed to get product by ID", "error", parsedErr)
		return nil, parsedErr
	}
	return prd, nil
}

func (s *Service) GetProductByExternalID(ctx context.Context, externalID string) (*models.Product, error) {
	prd, err := s.storage.Products().GetByExternalID(ctx, pgtypeutils.EncodeText(&externalID))
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	return prd, nil
}

func (s *Service) UpdateProduct(ctx context.Context, d dto.UpdateProductDTO) (*models.Product, *models.ProductVariant, error) {
	productParams := repository_products.UpdateParams{
		ID:             d.ID,
		ManufacturerID: d.ManufacturerID,
		Model:          d.Model,
		Sku:            pgtypeutils.EncodeText(d.Sku),
		Upc:            pgtypeutils.EncodeText(d.Upc),
		Ean:            pgtypeutils.EncodeText(d.Ean),
		Jan:            pgtypeutils.EncodeText(d.Jan),
		Isbn:           pgtypeutils.EncodeText(d.Isbn),
		Mpn:            pgtypeutils.EncodeText(d.Mpn),
		Location:       pgtypeutils.EncodeText(d.Location),
		Quantity:       d.Quantity,
		StockStatus:    d.StockStatus,
		Price:          d.Price,
		Weight:         d.Weight,
		Length:         d.Length,
		Width:          d.Width,
		Height:         d.Height,
		Subtract:       d.Subtract,
		Minimum:        d.Minimum,
		SortOrder:      d.SortOrder,
		IsEnable:       d.IsEnable,
	}

	var prd *models.Product
	var variant *models.ProductVariant
	var err error

	err = repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		prd, err = s.storage.Products(repository.WithTx(tx)).Update(ctx, productParams)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to update product", parsedErr)
			return parsedErr
		}

		variantParams := repository_product_variants.UpdateParams{
			ID:              d.Variant.ID,
			ProductID:       prd.ID,
			CategoryID:      d.Variant.CategoryID,
			Name:            d.Variant.Name,
			Slug:            d.Variant.Slug,
			Description:     pgtypeutils.EncodeText(d.Variant.Description),
			MetaTitle:       pgtypeutils.EncodeText(d.Variant.MetaTitle),
			MetaH1:          pgtypeutils.EncodeText(d.Variant.MetaH1),
			MetaDescription: pgtypeutils.EncodeText(d.Variant.MetaDescription),
			MetaKeyword:     pgtypeutils.EncodeText(d.Variant.MetaKeyword),
			Image:           pgtypeutils.EncodeText(d.Variant.Image),
			SortOrder:       d.Variant.SortOrder,
			IsEnable:        d.Variant.IsEnable,
			Viewed:          0,
		}
		variant, err = s.storage.ProductVariants(repository.WithTx(tx)).Update(ctx, variantParams)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to update product variant", parsedErr)
			return parsedErr
		}

		err = s.storage.Products(repository.WithTx(tx)).DeleteProductMedia(ctx, prd.ID)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to delete product media", parsedErr)
			return parsedErr
		}
		for _, mediaID := range d.MediaIDs {
			err := s.storage.Products(repository.WithTx(tx)).CreateProductMedia(ctx, repository_products.CreateProductMediaParams{
				ProductID: prd.ID,
				MediaID:   *mediaID,
			})
			if err != nil {
				parsedErr := pgerror.ParseError(err)
				s.logger.Error("failed to create product media", parsedErr)
				return parsedErr
			}
		}

		err = s.IndexProduct(ctx, prd)
		if err != nil {
			s.logger.Error("failed to index product", err)
			return err
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return prd, variant, nil
}

func (s *Service) GetProductWithMediaByID(ctx context.Context, id uuid.UUID) (*dto.ProductWithMediaDTO, error) {
	product, err := s.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	variants, err := s.storage.ProductVariants().GetByProductID(ctx, id)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	media, err := s.storage.Products().GetMediaByProductID(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get product media", err)
		return nil, parsedErr
	}

	var primaryVariant *models.ProductVariant
	if len(variants) > 0 {
		primaryVariant = variants[0]
	}

	return &dto.ProductWithMediaDTO{
		Product: product,
		Variant: primaryVariant,
		Medium:  media,
	}, nil
}

func (s *Service) SyncProductAttributes(ctx context.Context, d dto.SyncAttributeProductDTO) error {
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.storage.ProductAttributeValues(repository.WithTx(tx)).RemoveAll(ctx, d.ProductID)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to remove product attribute values", parsedErr)
			return parsedErr
		}

		if len(d.AttributeValueIDs) == 0 {
			return nil
		}

		err = s.storage.ProductAttributeValues(repository.WithTx(tx)).AddBatch(ctx, repository_product_attribute_values.AddBatchParams{
			ProductID:        d.ProductID,
			AttributeValueID: d.AttributeValueIDs,
		})
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to add product attribute values", parsedErr)
			return parsedErr
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetProductAttributesByID(ctx context.Context, id uuid.UUID) ([]*dto.AttributeGroupWithValuesDTO, error) {
	rawAttributes, err := s.storage.ProductAttributeValues().GetByProductID(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get product attributes", parsedErr)
		return nil, parsedErr
	}

	return mapper.MapProductAttributesToGroupedDTO(rawAttributes), nil
}

func (s *Service) UpsertProductByExternalID(ctx context.Context, externalID string, d dto.ProductUpsertDTO) (*models.Product, error) {
	existing, err := s.storage.Products().GetByExternalID(ctx, pgtypeutils.EncodeText(&externalID))
	if err != nil {
		var notFound *pgerror.NotFoundError
		if !errors.As(pgerror.ParseError(err), &notFound) {
			return nil, pgerror.ParseError(err)
		}
		// товар не найден — создаём только запись товара без варианта
		return s.storage.Products().Create(ctx, repository_products.CreateParams{
			ExternalID:     pgtypeutils.EncodeText(&externalID),
			ManufacturerID: d.ManufacturerID,
			Model:          d.Model,
			Sku:            pgtypeutils.EncodeText(d.Sku),
			Quantity:       d.Quantity,
			StockStatus:    d.StockStatus,
			Price:          d.Price,
			IsEnable:       d.IsEnable,
		})
	}

	// товар найден — обновляем только поля товара
	prd, err := s.storage.Products().Update(ctx, repository_products.UpdateParams{
		ID:             existing.ID,
		ExternalID:     pgtypeutils.EncodeText(&externalID),
		ManufacturerID: d.ManufacturerID,
		Model:          d.Model,
		Sku:            pgtypeutils.EncodeText(d.Sku),
		Quantity:       d.Quantity,
		StockStatus:    d.StockStatus,
		Price:          d.Price,
		IsEnable:       d.IsEnable,
	})
	if err != nil {
		return nil, pgerror.ParseError(err)
	}
	return prd, nil
}
