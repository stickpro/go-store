package collections

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/dto/mapper"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_collections"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type ICollectionsService interface {
	CreateCollection(ctx context.Context, d dto.CreateCollectionDTO) (*models.Collection, error)
	GetCollectionByID(ctx context.Context, id uuid.UUID) (*dto.WithProductsCollectionDTO, error)
	GetCollectionBySlug(ctx context.Context, slug string) (*dto.WithProductsCollectionDTO, error)
	GetCollectionsWithPagination(ctx context.Context, d dto.GetCollectionDTO) (*base.FindResponseWithFullPagination[*models.Collection], error)
	UpdateCollection(ctx context.Context, d dto.UpdateCollectionDTO) (*models.Collection, error)
	DeleteCollection(ctx context.Context, id uuid.UUID) error
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

func (s *Service) CreateCollection(ctx context.Context, d dto.CreateCollectionDTO) (*models.Collection, error) {
	params := repository_collections.CreateParams{
		Name: d.Name,
		Slug: d.Slug,
	}
	if d.Description != nil {
		params.Description = pgtypeutils.EncodeText(d.Description)
	}

	var col *models.Collection
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		collection, err := s.storage.Collections().Create(ctx, params)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.l.Error("failed to create collection", "error", parsedErr)
			return parsedErr
		}
		col = collection
		err = s.updateProductToCollection(ctx, collection, d.ProductIDs, tx)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return col, nil
}

func (s *Service) GetCollectionByID(ctx context.Context, id uuid.UUID) (*dto.WithProductsCollectionDTO, error) {
	rows, err := s.storage.Collections().GetCollectionWithProductsByID(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("failed to get collection by ID ", parsedErr)
		return nil, parsedErr
	}
	d := mapper.MapCollectionToDTO(rows)
	return d, nil
}
func (s *Service) GetCollectionBySlug(ctx context.Context, slug string) (*dto.WithProductsCollectionDTO, error) {
	rows, err := s.storage.Collections().GetCollectionWithProductsBySlug(ctx, slug)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("failed to get collection by ID", "error", parsedErr)
		return nil, parsedErr
	}
	d := mapper.MapCollectionBySlugToDTO(rows)
	return d, nil
}

func (s *Service) GetCollectionsWithPagination(ctx context.Context, d dto.GetCollectionDTO) (*base.FindResponseWithFullPagination[*models.Collection], error) {
	commonParams := base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}
	commonParams.WithDeleted = false

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

func (s *Service) UpdateCollection(ctx context.Context, d dto.UpdateCollectionDTO) (*models.Collection, error) {
	params := repository_collections.UpdateParams{
		ID:          d.ID,
		Name:        d.Name,
		Slug:        d.Slug,
		Description: pgtypeutils.EncodeText(d.Description),
	}
	var col *models.Collection
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		collection, err := s.storage.Collections().Update(ctx, params)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.l.Error("failed to update collection", "error", parsedErr)
			return parsedErr
		}
		col = collection
		err = s.updateProductToCollection(ctx, collection, d.ProductIDs, tx)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return col, nil
}

func (s *Service) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.storage.Collections().Delete(ctx, id)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.l.Error("failed to delete collection", "error", parsedErr)
			return parsedErr
		}
		err = s.storage.Collections(repository.WithTx(tx)).DeleteProductsFromCollection(ctx, id)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) updateProductToCollection(
	ctx context.Context,
	collection *models.Collection,
	productIDs []uuid.UUID,
	dbTx pgx.Tx,
) error {
	err := s.storage.Collections(repository.WithTx(dbTx)).DeleteProductsFromCollection(ctx, collection.ID)
	if err != nil {
		return err
	}
	if len(productIDs) == 0 {
		return nil
	}
	err = s.storage.Collections(repository.WithTx(dbTx)).AddProductsToCollection(ctx, repository_collections.AddProductsToCollectionParams{
		CollectionID: collection.ID,
		ProductIds:   productIDs,
	})
	if err != nil {
		return err
	}
	return nil
}
