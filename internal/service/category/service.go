package category

import (
	"context"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_categories"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type ICategoryService interface {
	CreateCategory(ctx context.Context, dto CreateDTO) (*models.Category, error)
	GetCategoryWithPagination(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*repository_categories.FindRow], error)
	GetCategoryById(ctx context.Context, id uuid.UUID) (*models.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
	UpdateCategory(ctx context.Context, dto UpdateDTO) (*models.Category, error)
}

type Service struct {
	cfg     *config.Config
	logger  logger.Logger
	storage storage.IStorage
}

func (s *Service) CreateCategory(ctx context.Context, dto CreateDTO) (*models.Category, error) {
	params := repository_categories.CreateParams{
		ParentID:        dto.ParentID,
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
	cat, err := s.storage.Categories().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create category", "error", parsedErr)
		return nil, parsedErr
	}
	return cat, nil
}

func (s *Service) GetCategoryById(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	cat, err := s.storage.Categories().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *Service) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	cat, err := s.storage.Categories().GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *Service) GetCategoryWithPagination(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*repository_categories.FindRow], error) {
	commonParams := base.NewCommonFindParams()
	if dto.PageSize != nil {
		commonParams.PageSize = dto.PageSize
	}
	if dto.Page != nil {
		commonParams.Page = dto.Page
	}
	cats, err := s.storage.Categories().GetWithPaginate(ctx, repository_categories.CategoryWithPaginationParams{
		CommonFindParams: *commonParams,
	})

	if err != nil {
		return nil, err
	}
	return cats, nil
}

func (s *Service) UpdateCategory(ctx context.Context, dto UpdateDTO) (*models.Category, error) {
	params := repository_categories.UpdateParams{
		Name:            dto.Name,
		ParentID:        dto.ParentID,
		Slug:            dto.Slug,
		Description:     pgtypeutils.EncodeText(dto.Description),
		MetaTitle:       pgtypeutils.EncodeText(dto.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(dto.MetaH1),
		MetaKeyword:     pgtypeutils.EncodeText(dto.MetaKeyword),
		MetaDescription: pgtypeutils.EncodeText(dto.MetaDescription),
		IsEnable:        dto.IsEnable,
		ID:              dto.ID,
	}

	cat, err := s.storage.Categories().Update(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to update category", "error", parsedErr)
		return nil, parsedErr
	}
	return cat, nil
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		logger:  logger,
		storage: storage,
	}
}
