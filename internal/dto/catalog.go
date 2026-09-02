package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/storage/base"
)

// CategoryProductsSort enumerates the sort options accepted by the storefront catalog.
type CategoryProductsSort string

const (
	CategoryProductsSortDefault   CategoryProductsSort = ""
	CategoryProductsSortPriceAsc  CategoryProductsSort = "price_asc"
	CategoryProductsSortPriceDesc CategoryProductsSort = "price_desc"
	CategoryProductsSortNew       CategoryProductsSort = "new"
	CategoryProductsSortPopular   CategoryProductsSort = "popular"
	CategoryProductsSortName      CategoryProductsSort = "name"
)

// AttributeFilterDTO is a single attribute constraint: either a set of accepted values
// (select/text/boolean attributes) or a numeric [Min, Max] range (number attributes).
type AttributeFilterDTO struct {
	Slug   string
	Values []string
	Min    *float64
	Max    *float64
}

// CategoryProductsFilterDTO is the fully parsed request for the storefront category listing.
type CategoryProductsFilterDTO struct {
	CategorySlug    string
	Page            *uint64
	PageSize        *uint64
	Sort            CategoryProductsSort
	PriceMin        *decimal.Decimal
	PriceMax        *decimal.Decimal
	ManufacturerIDs []uuid.UUID
	StockStatuses   []string
	Attributes      []AttributeFilterDTO
	WithFacets      bool
}

// CategoryProductsResultDTO is the listing payload: page of variants plus facet data.
type CategoryProductsResultDTO struct {
	Items      []*EnrichedVariantDTO        `json:"items"`
	Pagination base.FullPagingData          `json:"pagination"`
	Facets     map[string]map[string]int64  `json:"facets,omitempty"`
	FacetStats map[string]CategoryFacetStat `json:"facet_stats,omitempty"` //nolint:tagliatelle
}

// CategoryFacetStat is the numeric range of a facet across the current result set.
type CategoryFacetStat struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// CategoryFiltersDTO is the full set of filters available for a category listing,
// with per-option product counts, computed over the category and its whole subtree.
type CategoryFiltersDTO struct {
	Price         *CategoryPriceRangeDTO       `json:"price,omitempty"`
	Manufacturers []CategoryFilterOptionDTO    `json:"manufacturers"`
	StockStatuses []CategoryFilterOptionDTO    `json:"stock_statuses"`
	Attributes    []CategoryAttributeFilterDTO `json:"attributes"`
}

type CategoryPriceRangeDTO struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type CategoryFilterOptionDTO struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type CategoryAttributeFilterDTO struct {
	Slug      string                    `json:"slug"`
	Name      string                    `json:"name"`
	Type      string                    `json:"type"` // select | number | boolean | text
	Unit      *string                   `json:"unit,omitempty"`
	GroupSlug string                    `json:"group_slug"`
	GroupName string                    `json:"group_name"`
	Options   []CategoryFilterOptionDTO `json:"options,omitempty"`
	Min       *float64                  `json:"min,omitempty"` // number attributes only
	Max       *float64                  `json:"max,omitempty"` // number attributes only
}
