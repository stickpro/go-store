package collections

import (
	"context"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_collections"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type ICollectionsService interface {
	CreateCollection(ctx context.Context, dto CreateDTO) (*models.Collection, error)
	GetCollectionByID(ctx context.Context, ID uuid.UUID) (*models.Collection, error)
	GetCollectionBySlug(ctx context.Context, slug string) (*models.Collection, error)
	GetCollectionsWithPagination(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*models.Collection], error)
	UpdateCollection(ctx context.Context, dto UpdateDTO) (*models.Collection, error)
	DeleteCollection(ctx context.Context, ID uuid.UUID) error
}

type Service struct {
	cfg     *config.Config
	l       logger.Logger
	storage storage.IStorage
}

func New(cfg *config.Config, log logger.Logger, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		l:       log,
		storage: storage,
	}
}

func (s *Service) CreateCollection(ctx context.Context, dto CreateDTO) (*models.Collection, error) {
	params := repository_collections.CreateParams{
		Name: dto.Name,
		Slug: dto.Slug,
	}
	if dto.Description != nil {
		params.Description = pgtypeutils.EncodeText(dto.Description)
	}

	collection, err := s.storage.Collections().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Error("failed to create collection", "error", parsedErr)
		return nil, parsedErr
	}
	return collection, nil
}

func (s *Service) GetCollectionByID(ctx context.Context, ID uuid.UUID) (*models.Collection, error) {
	collection, err := s.storage.Collections().Get(ctx, ID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("failed to get collection by ID ", parsedErr)
		return nil, parsedErr
	}
	return collection, nil
}
func (s *Service) GetCollectionBySlug(ctx context.Context, slug string) (*models.Collection, error) {
	collection, err := s.storage.Collections().GetBySlug(ctx, slug)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("failed to get collection by ID", "error", parsedErr)
		return nil, parsedErr
	}
	return collection, nil
}

func (s *Service) GetCollectionsWithPagination(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*models.Collection], error) {
	commonParams := base.NewCommonFindParams()
	if dto.PageSize != nil {
		commonParams.PageSize = dto.PageSize
	}
	if dto.Page != nil {
		commonParams.Page = dto.Page
	}

	collections, err := s.storage.Collections().GetWithPaginate(ctx, repository_collections.CollectionWithPaginationParams{
		CommonFindParams: *commonParams,
	})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("failed to get collections with pagination", "error", parsedErr)
		return nil, parsedErr
	}
	return collections, nil
}

func (s *Service) UpdateCollection(ctx context.Context, dto UpdateDTO) (*models.Collection, error) {
	params := repository_collections.UpdateParams{
		ID:          dto.ID,
		Name:        dto.Name,
		Slug:        dto.Slug,
		Description: pgtypeutils.EncodeText(dto.Description),
	}

	collection, err := s.storage.Collections().Update(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Error("failed to update collection", "error", parsedErr)
		return nil, parsedErr
	}
	return collection, nil
}

func (s *Service) DeleteCollection(ctx context.Context, ID uuid.UUID) error {
	err := s.storage.Collections().Delete(ctx, ID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Error("failed to delete collection", "error", parsedErr)
		return parsedErr
	}
	return nil
}
