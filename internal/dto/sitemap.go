package dto

import "time"

// SitemapEntryDTO is one URL row of a sitemap feed: the frontend maps Slug to a
// route and uses UpdatedAt as <lastmod>. UpdatedAt is nil when the source row
// has no updated_at.
type SitemapEntryDTO struct {
	Slug      string
	UpdatedAt *time.Time
}
