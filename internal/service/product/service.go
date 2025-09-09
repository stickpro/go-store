package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type IProductService interface {
	CreateProduct(ctx context.Context, d dto.CreateProductDTO) (*models.Product, error)
	GetProductWithPagination(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*repository_products.FindRow], error)
	GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetProductWithMediumByID(ctx context.Context, id uuid.UUID) (*dto.WithMediumProductDTO, error)
	UpdateProduct(ctx context.Context, d dto.UpdateProductDTO) (*models.Product, error)
	//
	SyncProductAttributes(ctx context.Context, d dto.SyncAttributeProductDTO) error
	// CreateProductIndex Indexing
	CreateProductIndex(ctx context.Context, reindex bool) error
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

func (s *Service) CreateProduct(ctx context.Context, d dto.CreateProductDTO) (*models.Product, error) {
	params := repository_products.CreateParams{
		Name:            d.Name,
		Model:           d.Model,
		Slug:            d.Slug,
		Description:     pgtypeutils.EncodeText(d.Description),
		MetaTitle:       pgtypeutils.EncodeText(d.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(d.MetaH1),
		MetaKeyword:     pgtypeutils.EncodeText(d.MetaKeyword),
		MetaDescription: pgtypeutils.EncodeText(d.MetaDescription),
		Sku:             pgtypeutils.EncodeText(d.Sku),
		Upc:             pgtypeutils.EncodeText(d.Upc),
		Ean:             pgtypeutils.EncodeText(d.Ean),
		Jan:             pgtypeutils.EncodeText(d.Jan),
		Isbn:            pgtypeutils.EncodeText(d.Isbn),
		Mpn:             pgtypeutils.EncodeText(d.Mpn),
		Location:        pgtypeutils.EncodeText(d.Location),
		Quantity:        d.Quantity,
		StockStatus:     d.StockStatus,
		Image:           pgtypeutils.EncodeText(d.Image),
		ManufacturerID:  d.ManufacturerID,
		Price:           d.Price,
		Weight:          d.Weight,
		Length:          d.Length,
		Width:           d.Width,
		Height:          d.Height,
		Subtract:        d.Subtract,
		Minimum:         d.Minimum,
		SortOrder:       d.SortOrder,
		IsEnable:        d.IsEnable,
		Viewed:          0,
	}
	prd := &models.Product{}
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		prd, err := s.storage.Products(repository.WithTx(tx)).Create(ctx, params)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to create product", parsedErr)
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
		return nil, err
	}

	return prd, nil
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

func (s *Service) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	prd, err := s.storage.Products().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return prd, nil
}

func (s *Service) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	prd, err := s.storage.Products().GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return prd, nil
}

func (s *Service) UpdateProduct(ctx context.Context, d dto.UpdateProductDTO) (*models.Product, error) {
	params := repository_products.UpdateParams{
		ID:              d.ID,
		Name:            d.Name,
		Model:           d.Model,
		Slug:            d.Slug,
		Description:     pgtypeutils.EncodeText(d.Description),
		MetaTitle:       pgtypeutils.EncodeText(d.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(d.MetaH1),
		MetaKeyword:     pgtypeutils.EncodeText(d.MetaKeyword),
		MetaDescription: pgtypeutils.EncodeText(d.MetaDescription),
		Sku:             pgtypeutils.EncodeText(d.Sku),
		Upc:             pgtypeutils.EncodeText(d.Upc),
		Ean:             pgtypeutils.EncodeText(d.Ean),
		Jan:             pgtypeutils.EncodeText(d.Jan),
		Isbn:            pgtypeutils.EncodeText(d.Isbn),
		Mpn:             pgtypeutils.EncodeText(d.Mpn),
		Location:        pgtypeutils.EncodeText(d.Location),
		Quantity:        d.Quantity,
		StockStatus:     d.StockStatus,
		Image:           pgtypeutils.EncodeText(d.Image),
		ManufacturerID:  d.ManufacturerID,
		Price:           d.Price,
		Weight:          d.Weight,
		Length:          d.Length,
		Width:           d.Width,
		Height:          d.Height,
		Subtract:        d.Subtract,
		Minimum:         d.Minimum,
		SortOrder:       d.SortOrder,
		IsEnable:        d.IsEnable,
	}
	prd := &models.Product{}
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		prd, err := s.storage.Products(repository.WithTx(tx)).Update(ctx, params)
		if err != nil {
			if err != nil {
				parsedErr := pgerror.ParseError(err)
				s.logger.Error("failed to update product ", parsedErr)
				return parsedErr
			}
		}
		err = s.storage.Products(repository.WithTx(tx)).DeleteProductMedia(ctx, prd.ID)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to delete product", parsedErr)
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
		return nil, err
	}

	return prd, nil
}

func (s *Service) GetProductWithMediumByID(ctx context.Context, id uuid.UUID) (*dto.WithMediumProductDTO, error) {
	product, err := s.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	media, err := s.storage.Products().GetMediaByProductID(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get product media", err)
		return nil, parsedErr
	}

	return &dto.WithMediumProductDTO{
		Product: product,
		Medium:  media,
	}, nil
}

func (s *Service) SyncProductAttributes(ctx context.Context, d dto.SyncAttributeProductDTO) error {
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.storage.Products(repository.WithTx(tx)).DeleteAttributesFromProduct(ctx, d.ProductID)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to delete attributes from product", parsedErr)
			return parsedErr
		}
		if len(d.AttributeIDs) == 0 {
			return nil
		}

		err = s.storage.Products(repository.WithTx(tx)).AddAttributesToProduct(ctx, repository_products.AddAttributesToProductParams{
			ProductID:    d.ProductID,
			AttributeIds: d.AttributeIDs,
		})

		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to add attributes to product", parsedErr)
			return parsedErr
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}
