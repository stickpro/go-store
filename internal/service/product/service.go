package product

import (
	"context"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type IProduct interface {
	CreateProduct(ctx context.Context, dto CreateDTO) (*models.Product, error)
	GetProductWithPagination(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*repository_products.FindRow], error)
	GetProductById(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	UpdateProduct(ctx context.Context, dto UpdateDTO) (*models.Product, error)
}

type Service struct {
	cfg     *config.Config
	logger  logger.Logger
	storage storage.IStorage
}

func (s Service) CreateProduct(ctx context.Context, dto CreateDTO) (*models.Product, error) {
	params := repository_products.CreateParams{
		Name:            dto.Name,
		Model:           dto.Model,
		Slug:            dto.Slug,
		Description:     pgtypeutils.EncodeText(dto.Description),
		MetaTitle:       pgtypeutils.EncodeText(dto.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(dto.MetaH1),
		MetaKeyword:     pgtypeutils.EncodeText(dto.MetaKeyword),
		MetaDescription: pgtypeutils.EncodeText(dto.MetaDescription),
		Sku:             pgtypeutils.EncodeText(dto.Sku),
		Upc:             pgtypeutils.EncodeText(dto.Upc),
		Ean:             pgtypeutils.EncodeText(dto.Ean),
		Jan:             pgtypeutils.EncodeText(dto.Jan),
		Isbn:            pgtypeutils.EncodeText(dto.Isbn),
		Mpn:             pgtypeutils.EncodeText(dto.Mpn),
		Location:        pgtypeutils.EncodeText(dto.Location),
		Quantity:        dto.Quantity,
		StockStatus:     dto.StockStatus,
		Image:           pgtypeutils.EncodeText(dto.Image),
		ManufacturerID:  dto.ManufacturerID,
		Price:           dto.Price,
		Weight:          dto.Weight,
		Length:          dto.Length,
		Width:           dto.Width,
		Height:          dto.Height,
		Subtract:        dto.Subtract,
		Minimum:         dto.Minimum,
		SortOrder:       dto.SortOrder,
		IsEnable:        dto.IsEnable,
		Viewed:          0,
	}

	prd, err := s.storage.Products().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create product", "error", parsedErr)
		return nil, parsedErr
	}
	return prd, nil

}

func (s Service) GetProductWithPagination(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*repository_products.FindRow], error) {
	commonParams := base.NewCommonFindParams()
	if dto.PageSize != nil {
		commonParams.PageSize = dto.PageSize
	}
	if dto.Page != nil {
		commonParams.Page = dto.Page
	}

	prds, err := s.storage.Products().GetWithPaginate(ctx, repository_products.ProductsWithPaginationParams{
		CommonFindParams: *commonParams,
	})
	if err != nil {
		return nil, err
	}
	return prds, nil
}

func (s Service) GetProductById(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	prd, err := s.storage.Products().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return prd, nil
}

func (s Service) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	prd, err := s.storage.Products().GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return prd, nil
}

func (s Service) UpdateProduct(ctx context.Context, dto UpdateDTO) (*models.Product, error) {
	params := repository_products.UpdateParams{
		ID:              dto.ID,
		Name:            dto.Name,
		Model:           dto.Model,
		Slug:            dto.Slug,
		Description:     pgtypeutils.EncodeText(dto.Description),
		MetaTitle:       pgtypeutils.EncodeText(dto.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(dto.MetaH1),
		MetaKeyword:     pgtypeutils.EncodeText(dto.MetaKeyword),
		MetaDescription: pgtypeutils.EncodeText(dto.MetaDescription),
		Sku:             pgtypeutils.EncodeText(dto.Sku),
		Upc:             pgtypeutils.EncodeText(dto.Upc),
		Ean:             pgtypeutils.EncodeText(dto.Ean),
		Jan:             pgtypeutils.EncodeText(dto.Jan),
		Isbn:            pgtypeutils.EncodeText(dto.Isbn),
		Mpn:             pgtypeutils.EncodeText(dto.Mpn),
		Location:        pgtypeutils.EncodeText(dto.Location),
		Quantity:        dto.Quantity,
		StockStatus:     dto.StockStatus,
		Image:           pgtypeutils.EncodeText(dto.Image),
		ManufacturerID:  dto.ManufacturerID,
		Price:           dto.Price,
		Weight:          dto.Weight,
		Length:          dto.Length,
		Width:           dto.Width,
		Height:          dto.Height,
		Subtract:        dto.Subtract,
		Minimum:         dto.Minimum,
		SortOrder:       dto.SortOrder,
		IsEnable:        dto.IsEnable,
	}

	prd, err := s.storage.Products().Update(ctx, params)
	if err != nil {
		return nil, err
	}
	return prd, nil
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		logger:  logger,
		storage: storage,
	}
}
