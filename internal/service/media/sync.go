package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_media"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

// SyncProductImages downloads new images, reuses existing ones, removes stale ones.
// Images uploaded via admin UI (source_url IS NULL) are never touched.
// TODO main image set for product or product_variant?
func (s Service) SyncProductImages(ctx context.Context, productID uuid.UUID, imageMain *string, images []string) error {
	// Build ordered URL list: main first, then the rest (deduplicated)
	seen := make(map[string]struct{})
	var orderedURLs []string
	for _, u := range append(urlSlice(imageMain), images...) {
		if _, ok := seen[u]; !ok {
			seen[u] = struct{}{}
			orderedURLs = append(orderedURLs, u)
		}
	}

	// Get current Kafka-managed media for this product (source_url IS NOT NULL)
	currentMedia, err := s.storage.Products().GetMediaByProductID(ctx, productID)
	if err != nil {
		return fmt.Errorf("get product media: %w", pgerror.ParseError(err))
	}
	currentByURL := make(map[string]*models.Medium, len(currentMedia))
	for _, m := range currentMedia {
		if m.SourceUrl.Valid {
			currentByURL[m.SourceUrl.String] = m
		}
	}

	// Resolve each URL to a media record in parallel (reuse or download).
	// Results are collected in order to preserve sort_order.
	type result struct {
		mediaID uuid.UUID
		err     error
	}
	results := make([]result, len(orderedURLs))
	var wg sync.WaitGroup
	for i, rawURL := range orderedURLs {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			m, err := s.resolveImage(ctx, u)
			if err != nil {
				results[idx] = result{err: err}
				return
			}
			results[idx] = result{mediaID: m.ID}
		}(i, rawURL)
	}
	wg.Wait()

	newMediaIDs := make([]uuid.UUID, 0, len(orderedURLs))
	for i, r := range results {
		if r.err != nil {
			s.l.Errorw("sync product images: resolve", "url", orderedURLs[i], "error", r.err)
			continue
		}
		newMediaIDs = append(newMediaIDs, r.mediaID)
	}

	// Determine which Kafka-managed media should be removed from product_media
	newSet := make(map[uuid.UUID]struct{}, len(newMediaIDs))
	for _, id := range newMediaIDs {
		newSet[id] = struct{}{}
	}
	var toRemove []uuid.UUID
	for _, m := range currentMedia {
		if !m.SourceUrl.Valid {
			continue // admin-managed, skip
		}
		if _, keep := newSet[m.ID]; !keep {
			toRemove = append(toRemove, m.ID)
		}
	}

	return repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		repo := s.storage.Products(repository.WithTx(tx))

		if len(toRemove) > 0 {
			if err := repo.DeleteProductMediaByMediaIDs(ctx, repository_products.DeleteProductMediaByMediaIDsParams{
				ProductID: productID,
				Column2:   toRemove,
			}); err != nil {
				return pgerror.ParseError(err)
			}
		}

		for i, mediaID := range newMediaIDs {
			if err := repo.CreateProductMedia(ctx, repository_products.CreateProductMediaParams{
				ProductID: productID,
				MediaID:   mediaID,
				SortOrder: int32(i), //nolint:gosec
			}); err != nil {
				return pgerror.ParseError(err)
			}
		}
		return nil
	})
}

// resolveImage returns an existing media record for the URL or downloads and creates a new one.
func (s Service) resolveImage(ctx context.Context, rawURL string) (*models.Medium, error) {
	existing, err := s.storage.Media().GetBySourceURL(ctx, pgtype.Text{String: rawURL, Valid: true})
	if err == nil {
		return existing, nil
	}
	if !isNotFound(err) {
		return nil, fmt.Errorf("lookup by source url: %w", pgerror.ParseError(err))
	}

	data, mimeType, err := downloadImage(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", rawURL, err)
	}

	fileName := sanitizeFileName(rawURL)
	storagePath := "public/images/" + fileName
	fPath, err := s.objectStorage.Save(ctx, storagePath, data)
	if err != nil {
		return nil, fmt.Errorf("save to storage: %w", err)
	}

	medium, err := s.storage.Media().CreateWithSourceURL(ctx, repository_media.CreateWithSourceURLParams{
		Name:      fileName,
		Path:      fPath,
		FileName:  fileName,
		MimeType:  mimeType,
		DiskType:  s.cfg.FileStorage.Type,
		Size:      int64(len(data)),
		SourceUrl: pgtype.Text{String: rawURL, Valid: true},
	})
	if err != nil {
		_ = s.objectStorage.Delete(ctx, fPath)
		return nil, fmt.Errorf("create media record: %w", pgerror.ParseError(err))
	}
	return medium, nil
}

func downloadImage(ctx context.Context, rawURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	mimeType := http.DetectContentType(data)
	return data, mimeType, nil
}

func sanitizeFileName(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return uuid.New().String()
	}
	base := path.Base(parsed.Path)
	base = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, base)
	if base == "" || base == "." {
		return uuid.New().String()
	}
	return base
}

func urlSlice(s *string) []string {
	if s == nil {
		return nil
	}
	return []string{*s}
}

func isNotFound(err error) bool {
	var nf *pgerror.NotFoundError
	return errors.As(pgerror.ParseError(err), &nf)
}
