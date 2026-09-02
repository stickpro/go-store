package media

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/singleflight"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_media"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/imageprocessor"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/object_storage"
)

type IMediaService interface {
	Save(ctx context.Context, dto SaveMediumDTO) (*models.Medium, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SyncProductImages(ctx context.Context, productID uuid.UUID, imageMain *string, images []string) error
	// EnsureImageVariant returns the fs path of the "<base>_<size>.<ext>" resize, generating it on first use.
	EnsureImageVariant(ctx context.Context, fileName string) (string, error)
	Image(m *models.Medium, alt string) dto.ImageDTO
	Images(media []*models.Medium, alt string) []dto.ImageDTO
	PresetNames() []string
}

type Service struct {
	cfg           *config.Config
	l             logger.Logger
	objectStorage object_storage.IObjectStorage
	storage       storage.IStorage

	imgProcessor     imageprocessor.Processor
	imgPresets       map[string]imgPreset
	imgPresetsBySize map[int]imgPreset
	imgPresetOrder   []string
	imgJPEGQuality   int
	imgWebPQuality   int
	imgGroup         *singleflight.Group
}

func New(cfg *config.Config, l logger.Logger, st storage.IStorage) *Service {
	localStorage := object_storage.New(cfg.FileStorage.Path)
	byName, bySize, order := buildImagePresets(cfg.Images)
	jpegQ := cfg.Images.JpegQuality
	if jpegQ <= 0 {
		jpegQ = 82
	}
	webpQ := cfg.Images.WebpQuality
	if webpQ <= 0 {
		webpQ = 80
	}
	return &Service{
		cfg:              cfg,
		l:                l,
		objectStorage:    localStorage,
		storage:          st,
		imgProcessor:     imageprocessor.New(cfg.Images.MaxSourcePixels),
		imgPresets:       byName,
		imgPresetsBySize: bySize,
		imgPresetOrder:   order,
		imgJPEGQuality:   jpegQ,
		imgWebPQuality:   webpQ,
		imgGroup:         &singleflight.Group{},
	}
}

func (s *Service) Save(ctx context.Context, d SaveMediumDTO) (*models.Medium, error) {
	var medium *models.Medium
	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		fPath, err := s.objectStorage.Save(ctx, d.Path, d.Data)
		if err != nil {
			return err
		}
		w, h := s.probeDimensions(d.Data)
		params := repository_media.CreateParams{
			Name:     d.Name,
			Path:     fPath,
			FileName: d.Name,
			MimeType: d.FileType,
			Size:     d.Size,
			DiskType: s.cfg.FileStorage.Type,
			Width:    w,
			Height:   h,
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

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	mediaInfo, err := s.storage.Media().Get(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("failed to get media by ID", "error", parsedErr)
		return parsedErr
	}
	if err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.storage.Media(repository.WithTx(tx)).Delete(ctx, id); err != nil {
			parsedErr := pgerror.ParseError(err)
			s.l.Error("failed to delete media", "error", parsedErr)
			return parsedErr
		}
		return nil
	}); err != nil {
		return err
	}

	// Row is gone; best-effort remove the original + any cached resize variants
	// (public/images/<base>_<size>.<webp|jpg>). Leftover files are harmless.
	if err := s.objectStorage.Delete(ctx, mediaInfo.Path); err != nil {
		s.l.Warn("failed to delete original file", "path", mediaInfo.Path, "error", err)
	}
	base := strings.TrimSuffix(path.Base(mediaInfo.Path), path.Ext(mediaInfo.Path))
	for _, preset := range s.imgPresets {
		for _, ext := range []string{".webp", ".jpg"} {
			p := path.Join(imagesDir, fmt.Sprintf("%s_%d%s", base, preset.Size, ext))
			if err := s.objectStorage.Delete(ctx, p); err != nil {
				s.l.Warn("failed to delete cached variant", "path", p, "error", err)
			}
		}
	}
	return nil
}
