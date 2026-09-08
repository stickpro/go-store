package sitemap_response

import (
	"time"

	"github.com/stickpro/go-store/internal/dto"
)

// SitemapEntry is one URL of a sitemap feed. UpdatedAt maps to <lastmod> and is
// null when the source row has no timestamp.
type SitemapEntry struct {
	Slug      string     `json:"slug"`
	UpdatedAt *time.Time `json:"updated_at"`
} //	@name	SitemapEntry

func NewFromDTOs(items []dto.SitemapEntryDTO) []SitemapEntry {
	out := make([]SitemapEntry, 0, len(items))
	for _, it := range items {
		out = append(out, SitemapEntry{Slug: it.Slug, UpdatedAt: it.UpdatedAt})
	}
	return out
}
