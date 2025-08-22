package manufacturer

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_manufacturers"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type IManufacturerService interface {
	CreateManufacturer(ctx context.Context, dto CreateDTO) (*models.Manufacturer, error)
	GetManufacturersWithPagination(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*models.Manufacturer], error)
	GetManufacturerByID(ctx context.Context, id uuid.UUID) (*models.Manufacturer, error)
	GetManufacturerBySlug(ctx context.Context, slug string) (*models.Manufacturer, error)
	UpdateManufacturer(ctx context.Context, dto UpdateDTO) (*models.Manufacturer, error)
}

type Service struct {
	cfg     *config.Config
	logger  logger.Logger
	storage storage.IStorage
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		logger:  logger,
		storage: storage,
	}
}

func (s *Service) CreateManufacturer(ctx context.Context, dto CreateDTO) (*models.Manufacturer, error) {
	params := repository_manufacturers.CreateParams{
		Name:            dto.Name,
		Slug:            dto.Slug,
		Description:     pgtypeutils.EncodeText(dto.Description),
		ImagePath:       pgtypeutils.EncodeText(dto.ImagePath),
		MetaTitle:       pgtypeutils.EncodeText(dto.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(dto.MetaH1),
		MetaKeyword:     pgtypeutils.EncodeText(dto.MetaKeyword),
		MetaDescription: pgtypeutils.EncodeText(dto.MetaDescription),
		IsEnable:        true,
	}
	mfc, err := s.storage.Manufacturers().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create category", "error", parsedErr)
		return nil, parsedErr
	}
	return mfc, nil
}

func (s *Service) GetManufacturersWithPagination(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*models.Manufacturer], error) {
	commonParams := base.NewCommonFindParams()
	if dto.PageSize != nil {
		commonParams.PageSize = dto.PageSize
	}
	if dto.Page != nil {
		commonParams.Page = dto.Page
	}
	mfrs, err := s.storage.Manufacturers().GetWithPaginate(ctx, repository_manufacturers.ManufacturersWithPaginationParams{
		CommonFindParams: *commonParams,
	})
	if err != nil {
		return nil, err
	}
	return mfrs, nil
}

func (s *Service) GetManufacturerByID(ctx context.Context, id uuid.UUID) (*models.Manufacturer, error) {
	mfc, err := s.storage.Manufacturers().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mfc, nil
}

func (s *Service) GetManufacturerBySlug(ctx context.Context, slug string) (*models.Manufacturer, error) {
	mfc, err := s.storage.Manufacturers().GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return mfc, nil
}

func (s *Service) UpdateManufacturer(ctx context.Context, dto UpdateDTO) (*models.Manufacturer, error) {
	params := repository_manufacturers.UpdateParams{
		Name:            dto.Name,
		Slug:            dto.Slug,
		Description:     pgtypeutils.EncodeText(dto.Description),
		ImagePath:       pgtypeutils.EncodeText(dto.ImagePath),
		MetaTitle:       pgtypeutils.EncodeText(dto.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(dto.MetaH1),
		MetaKeyword:     pgtypeutils.EncodeText(dto.MetaKeyword),
		MetaDescription: pgtypeutils.EncodeText(dto.MetaDescription),
		IsEnable:        dto.IsEnable,
		ID:              dto.ID,
	}
	mfc, err := s.storage.Manufacturers().Update(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to update manufacturer", "error", parsedErr)
		return nil, parsedErr
	}
	return mfc, nil
}
