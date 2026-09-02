package product

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/key_value"
)

const filterableAttributesCacheTTL = 10 * time.Minute

// filterableAttr is the trimmed, cache-friendly view of a filterable attribute
// (only the fields the catalog needs; ordered by group name, sort order, name).
type filterableAttr struct {
	Slug      string  `json:"slug"`
	Name      string  `json:"name"`
	Type      string  `json:"type"` // select | number | boolean | text
	Unit      *string `json:"unit,omitempty"`
	GroupName string  `json:"group_name"`
	GroupSlug string  `json:"group_slug"`
}

// filterableAttributes returns every filterable attribute, cached in the key-value store.
// The set changes only when an admin edits attributes, so a short TTL plus explicit
// invalidation from the attribute service keeps it fresh without a query per request.
func (s *Service) filterableAttributes(ctx context.Context) ([]filterableAttr, error) {
	kv := s.storage.KeyValue()

	if raw, err := kv.Get(ctx, constant.CacheKeyFilterableAttributes); err == nil {
		var cached []filterableAttr
		if json.Unmarshal(raw.Bytes(), &cached) == nil {
			return cached, nil
		}
	} else if !errors.Is(err, key_value.ErrEntryNotFound) {
		s.logger.Warn("failed to read filterable attributes cache", "error", err)
	}

	rows, err := s.storage.Attributes().GetFilterableAttributes(ctx)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}

	attrs := make([]filterableAttr, 0, len(rows))
	for _, r := range rows {
		attrs = append(attrs, filterableAttr{
			Slug:      r.Slug,
			Name:      r.Name,
			Type:      r.Type,
			Unit:      pgtypeutils.DecodeText(r.Unit),
			GroupName: r.GroupName,
			GroupSlug: r.GroupSlug,
		})
	}

	if data, err := json.Marshal(attrs); err == nil {
		if err := kv.Set(ctx, constant.CacheKeyFilterableAttributes, string(data), filterableAttributesCacheTTL); err != nil {
			s.logger.Warn("failed to write filterable attributes cache", "error", err)
		}
	}

	return attrs, nil
}
