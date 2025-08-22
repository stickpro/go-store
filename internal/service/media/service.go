package media

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_media"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/object_storage"
)

type IMediaService interface {
	Save(ctx context.Context, dto SaveMediumDTO) (*models.Medium, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	cfg           *config.Config
	l             logger.Logger
	objectStorage object_storage.IObjectStorage
	storage       storage.IStorage
}

func New(cfg *config.Config, l logger.Logger, st storage.IStorage) *Service {
	localStorage := object_storage.New(cfg.FileStorage.Path)
	return &Service{
		cfg:           cfg,
		l:             l,
		objectStorage: localStorage,
		storage:       st,
	}
}

func (s Service) Save(ctx context.Context, dto SaveMediumDTO) (*models.Medium, error) {
	var medium *models.Medium
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		fPath, err := s.objectStorage.Save(ctx, dto.Path, dto.Data)
		if err != nil {
			return err
		}
		params := repository_media.CreateParams{
			Name:     dto.Name,
			Path:     fPath,
			FileName: dto.Name,
			MimeType: dto.FileType,
			Size:     dto.Size,
			DiskType: s.cfg.FileStorage.Type,
		}
		medium, err = s.storage.Media(repository.WithTx(tx)).Create(ctx, params)
		if err != nil {
			delErr := s.objectStorage.Delete(ctx, fPath)
			if delErr != nil {
				return delErr
			}
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return medium, nil
}

func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	mediaInfo, err := s.storage.Media().Get(ctx, id)

	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("failed to get media by ID", "error", parsedErr)
		return parsedErr
	}
	return repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err = s.storage.Media(repository.WithTx(tx)).Delete(ctx, id)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.l.Error("failed to delete media", "error", parsedErr)
			return parsedErr
		}
		err := s.objectStorage.Delete(ctx, mediaInfo.Path)
		if err != nil {
			return err
		}

		return nil
	})
}
