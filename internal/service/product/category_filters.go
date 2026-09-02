package product

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

// GetCategoryFilters returns every filter available for a category listing — price range,
// manufacturers, stock statuses and filterable attributes — each option carrying the number
// of matching products. Counts are computed from the search index over the category and its
// whole subtree, so they line up with SearchVariantsByCategory.
func (s *Service) GetCategoryFilters(ctx context.Context, categorySlug string) (*dto.CategoryFiltersDTO, error) {
	category, err := s.storage.Categories().GetBySlug(ctx, categorySlug)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}

	attrs, err := s.filterableAttributes(ctx)
	if err != nil {
		return nil, err
	}

	facetFields := make([]string, 0, len(attrs)+3)
	facetFields = append(facetFields, "manufacturer_id", "stock_status", "price")
	for _, a := range attrs {
		facetFields = append(facetFields, a.Slug)
	}

	res, err := s.searchService.SearchWithParams(constant.ProductVariantsIndex, searchtypes.SearchParams{
		Filter: "category_ids = " + meiliQuote(category.ID.String()) + " AND is_enable = true",
		Facets: facetFields,
		Limit:  0,
	})
	if err != nil {
		return nil, err
	}

	result := &dto.CategoryFiltersDTO{
		Manufacturers: s.manufacturerFilterOptions(ctx, res.Facets["manufacturer_id"]),
		StockStatuses: stockStatusFilterOptions(res.Facets["stock_status"]),
		Attributes:    attributeFilterOptions(attrs, res),
	}
	if stat, ok := res.FacetStats["price"]; ok {
		result.Price = &dto.CategoryPriceRangeDTO{Min: stat.Min, Max: stat.Max}
	}

	return result, nil
}

func (s *Service) manufacturerFilterOptions(ctx context.Context, dist map[string]int64) []dto.CategoryFilterOptionDTO {
	if len(dist) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(dist))
	for raw := range dist {
		if id, err := uuid.Parse(raw); err == nil {
			ids = append(ids, id)
		}
	}

	names := make(map[uuid.UUID]string, len(ids))
	if len(ids) > 0 {
		manufacturers, err := s.storage.Manufacturers().GetByIDs(ctx, ids)
		if err != nil {
			s.logger.Warn("failed to load manufacturers for category filters", "error", err)
		}
		for _, m := range manufacturers {
			names[m.ID] = m.Name
		}
	}

	opts := make([]dto.CategoryFilterOptionDTO, 0, len(dist))
	for raw, count := range dist {
		label := raw
		if id, err := uuid.Parse(raw); err == nil {
			if name, ok := names[id]; ok {
				label = name
			}
		}
		opts = append(opts, dto.CategoryFilterOptionDTO{Value: raw, Label: label, Count: count})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Label < opts[j].Label })
	return opts
}

func stockStatusFilterOptions(dist map[string]int64) []dto.CategoryFilterOptionDTO {
	labels := map[string]string{
		string(constant.InStock):    "In stock",
		string(constant.PreOrder):   "Pre-order",
		string(constant.OutOfStock): "Out of stock",
	}

	opts := make([]dto.CategoryFilterOptionDTO, 0, len(dist))
	for value, count := range dist {
		label := labels[value]
		if label == "" {
			label = value
		}
		opts = append(opts, dto.CategoryFilterOptionDTO{Value: value, Label: label, Count: count})
	}
	sort.Slice(opts, func(i, j int) bool { return opts[i].Value < opts[j].Value })
	return opts
}

func attributeFilterOptions(attrs []filterableAttr, res *searchtypes.SearchResult) []dto.CategoryAttributeFilterDTO {
	// attrs already come ordered by group name, attribute sort_order, attribute name.
	out := make([]dto.CategoryAttributeFilterDTO, 0, len(attrs))

	for _, a := range attrs {
		dist := res.Facets[a.Slug]
		stat, hasStat := res.FacetStats[a.Slug]

		if a.Type == "number" && !hasStat {
			continue
		}
		if a.Type != "number" && len(dist) == 0 {
			continue
		}

		f := dto.CategoryAttributeFilterDTO{
			Slug:      a.Slug,
			Name:      a.Name,
			Type:      a.Type,
			Unit:      a.Unit,
			GroupSlug: a.GroupSlug,
			GroupName: a.GroupName,
		}

		if a.Type == "number" {
			minV, maxV := stat.Min, stat.Max
			f.Min, f.Max = &minV, &maxV
		} else {
			f.Options = make([]dto.CategoryFilterOptionDTO, 0, len(dist))
			for value, count := range dist {
				f.Options = append(f.Options, dto.CategoryFilterOptionDTO{Value: value, Label: value, Count: count})
			}
			sort.Slice(f.Options, func(i, j int) bool { return f.Options[i].Value < f.Options[j].Value })
		}

		out = append(out, f)
	}

	return out
}
