package category

import (
	"context"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository/repository_categories"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type ICategory interface {
	CreateCategory(ctx context.Context, dto CreateDTO) (*models.Category, error)
	GetCategoryById(ctx context.Context, id uuid.UUID) (*models.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
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
		MetaTitle:       pgtypeutils.EncodeText(dto.MetaTitle),
		MetaH1:          pgtypeutils.EncodeText(dto.MetaH1),
		MetaKeyword:     pgtypeutils.EncodeText(dto.MetaKeyword),
		MetaDescription: pgtypeutils.EncodeText(dto.MetaDescription),
		IsEnable:        true,
	}
	cat, err := s.storage.Categories().Create(ctx, params)
	if err != nil {
		return nil, err
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

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		logger:  logger,
		storage: storage,
	}
}