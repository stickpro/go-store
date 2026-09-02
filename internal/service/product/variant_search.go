package product

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/search"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

const (
	defaultCategoryProductsPageSize = uint64(20)
	maxCategoryProductsPageSize     = uint64(100)
)

// SearchVariantsByCategory returns a paginated, filtered and faceted list of enabled product
// variants that belong to the category (found by slug) or any of its descendants, served from
// the search index. Index documents are pre-enriched with product-level fields, so this avoids
// the per-product database round-trips of the DB-backed path.
func (s *Service) SearchVariantsByCategory(
	ctx context.Context,
	d dto.CategoryProductsFilterDTO,
) (*dto.CategoryProductsResultDTO, error) {
	category, err := s.storage.Categories().GetBySlug(ctx, d.CategorySlug)
	if err != nil {
		return nil, pgerror.ParseError(err)
	}

	attrTypes, err := s.filterableAttributeTypes(ctx)
	if err != nil {
		return nil, err
	}

	page := uint64(1)
	if d.Page != nil && *d.Page > 0 {
		page = *d.Page
	}
	pageSize := defaultCategoryProductsPageSize
	if d.PageSize != nil && *d.PageSize > 0 {
		pageSize = *d.PageSize
	}
	if pageSize > maxCategoryProductsPageSize {
		pageSize = maxCategoryProductsPageSize
	}

	params := searchtypes.SearchParams{
		Filter: buildCategoryFilter(category.ID, d, attrTypes),
		Sort:   categorySortExpression(d.Sort),
		Limit:  int64(pageSize),              //nolint:gosec
		Offset: int64((page - 1) * pageSize), //nolint:gosec
	}
	if d.WithFacets {
		params.Facets = categoryFacetFields(attrTypes)
	}

	res, err := s.searchService.SearchWithParams(constant.ProductVariantsIndex, params)
	if err != nil {
		return nil, err
	}

	items, err := search.UnmarshalHits[*dto.EnrichedVariantDTO](res.Hits)
	if err != nil {
		return nil, err
	}

	total := uint64(res.TotalHits) //nolint:gosec
	lastPage := uint64(1)
	if total > 0 {
		lastPage = (total + pageSize - 1) / pageSize
	}

	result := &dto.CategoryProductsResultDTO{
		Items: items,
		Pagination: base.FullPagingData{
			Total:    total,
			PageSize: pageSize,
			Page:     page,
			LastPage: lastPage,
		},
	}
	if d.WithFacets {
		result.Facets = res.Facets
		if len(res.FacetStats) > 0 {
			result.FacetStats = make(map[string]dto.CategoryFacetStat, len(res.FacetStats))
			for name, stat := range res.FacetStats {
				result.FacetStats[name] = dto.CategoryFacetStat{Min: stat.Min, Max: stat.Max}
			}
		}
	}

	return result, nil
}

// filterableAttributeTypes returns a slug -> type ("select"|"number"|"boolean"|"text") map of
// every filterable attribute; used both to validate incoming attribute filters and to know how
// to render each one into a Meili filter clause.
func (s *Service) filterableAttributeTypes(ctx context.Context) (map[string]string, error) {
	attrs, err := s.filterableAttributes(ctx)
	if err != nil {
		return nil, err
	}

	types := make(map[string]string, len(attrs))
	for _, a := range attrs {
		types[a.Slug] = a.Type
	}
	return types, nil
}

func categoryFacetFields(attrTypes map[string]string) []string {
	fields := []string{"manufacturer_id", "stock_status", "price"}
	for slug := range attrTypes {
		fields = append(fields, slug)
	}
	return fields
}

func categorySortExpression(sort dto.CategoryProductsSort) []string {
	switch sort {
	case dto.CategoryProductsSortPriceAsc:
		return []string{"price:asc"}
	case dto.CategoryProductsSortPriceDesc:
		return []string{"price:desc"}
	case dto.CategoryProductsSortNew:
		return []string{"created_at:desc"}
	case dto.CategoryProductsSortPopular:
		return []string{"viewed:desc"}
	case dto.CategoryProductsSortName:
		return []string{"name:asc"}
	default:
		return []string{"sort_order:asc"}
	}
}

func buildCategoryFilter(
	categoryID uuid.UUID,
	d dto.CategoryProductsFilterDTO,
	attrTypes map[string]string,
) string {
	clauses := []string{
		"category_ids = " + meiliQuote(categoryID.String()),
		"is_enable = true",
	}

	if d.PriceMin != nil {
		clauses = append(clauses, "price >= "+d.PriceMin.String())
	}
	if d.PriceMax != nil {
		clauses = append(clauses, "price <= "+d.PriceMax.String())
	}

	if len(d.ManufacturerIDs) > 0 {
		ids := make([]string, len(d.ManufacturerIDs))
		for i, id := range d.ManufacturerIDs {
			ids[i] = id.String()
		}
		clauses = append(clauses, "manufacturer_id "+meiliInList(ids))
	}

	if len(d.StockStatuses) > 0 {
		clauses = append(clauses, "stock_status "+meiliInList(d.StockStatuses))
	}

	for _, af := range d.Attributes {
		attrType, ok := attrTypes[af.Slug]
		if !ok || !isSafeIdentifier(af.Slug) {
			continue // unknown or non-filterable attribute -> ignore
		}

		switch attrType {
		case "number":
			if af.Min != nil {
				clauses = append(clauses, fmt.Sprintf("%s >= %s", af.Slug, strconv.FormatFloat(*af.Min, 'f', -1, 64)))
			}
			if af.Max != nil {
				clauses = append(clauses, fmt.Sprintf("%s <= %s", af.Slug, strconv.FormatFloat(*af.Max, 'f', -1, 64)))
			}
			if nums := parseFloats(af.Values); len(nums) > 0 {
				clauses = append(clauses, af.Slug+" IN ["+strings.Join(nums, ", ")+"]")
			}
		case "boolean":
			if len(af.Values) > 0 {
				b := af.Values[0] == "true" || af.Values[0] == "1"
				clauses = append(clauses, fmt.Sprintf("%s = %t", af.Slug, b))
			}
		default: // select, text
			if len(af.Values) > 0 {
				clauses = append(clauses, af.Slug+" "+meiliInList(af.Values))
			}
		}
	}

	return strings.Join(clauses, " AND ")
}

// parseFloats keeps only the values that are valid numbers, formatted canonically.
func parseFloats(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			out = append(out, strconv.FormatFloat(f, 'f', -1, 64))
		}
	}
	return out
}

func meiliInList(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = meiliQuote(v)
	}
	return "IN [" + strings.Join(quoted, ", ") + "]"
}

// meiliQuote wraps a value in double quotes and escapes the characters that are
// significant inside a Meili filter string literal.
func meiliQuote(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		if r == '"' || r == '\\' {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}

// isSafeIdentifier guards against an attribute slug that could break out of its filter clause.
func isSafeIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}
