package product

import (
	"context"
	"fmt"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// GetSitemapEntries returns every enabled product-variant page (slug +
// updated_at) for the frontend sitemap generator. A variant is included only
// when both it and its parent product are enabled.
func (s *Service) GetSitemapEntries(ctx context.Context) ([]dto.SitemapEntryDTO, error) {
	rows, err := s.storage.ProductVariants().SitemapProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("product: sitemap entries: %w", err)
	}
	out := make([]dto.SitemapEntryDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.SitemapEntryDTO{Slug: r.Slug, UpdatedAt: pgtypeutils.DecodeTimePtr(r.UpdatedAt)})
	}
	return out, nil
}
