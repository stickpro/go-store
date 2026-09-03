package product_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

// VariantCardResponse is the single "product variant in a listing" contract, returned by
// storefront search, the category listing, related products and collections. Fields a given
// context does not carry (e.g. collections have no stock status) come back as their zero value.
type VariantCardResponse struct {
	ID             uuid.UUID        `json:"id"`
	ProductID      uuid.UUID        `json:"product_id"`
	CategoryID     uuid.NullUUID    `json:"category_id"`
	Name           string           `json:"name"`
	Slug           string           `json:"slug"`
	Model          string           `json:"model"`
	Description    *string          `json:"description"`
	PriceRetail    decimal.Decimal  `json:"price_retail"`
	PriceBusiness  decimal.Decimal  `json:"price_business"`
	PriceWholesale decimal.Decimal  `json:"price_wholesale"`
	StockStatus    string           `json:"stock_status"`
	ManufacturerID uuid.NullUUID    `json:"manufacturer_id"`
	Image          *models.ImageDTO `json:"image"`
	IsEnable       bool             `json:"is_enable"`
} //	@name	VariantCardResponse

// VariantListResponse is a paginated page of variant cards plus optional facet data.
type VariantListResponse struct {
	Items      []VariantCardResponse            `json:"items"`
	Pagination base.FullPagingData              `json:"pagination"`
	Facets     map[string]map[string]int64      `json:"facets,omitempty"`
	FacetStats map[string]dto.CategoryFacetStat `json:"facet_stats,omitempty"` //nolint:tagliatelle
} //	@name	VariantListResponse

func NewVariantCard(c *dto.VariantCardDTO) VariantCardResponse {
	return VariantCardResponse{
		ID:             c.ID,
		ProductID:      c.ProductID,
		CategoryID:     c.CategoryID,
		Name:           c.Name,
		Slug:           c.Slug,
		Model:          c.Model,
		Description:    c.Description,
		PriceRetail:    c.PriceRetail,
		PriceBusiness:  c.PriceBusiness,
		PriceWholesale: c.PriceWholesale,
		StockStatus:    c.StockStatus.String(),
		ManufacturerID: c.ManufacturerID,
		Image:          c.Image,
		IsEnable:       c.IsEnable,
	}
}

func NewVariantCards(cards []*dto.VariantCardDTO) []VariantCardResponse {
	out := make([]VariantCardResponse, 0, len(cards))
	for _, c := range cards {
		out = append(out, NewVariantCard(c))
	}
	return out
}

func NewVariantList(l *dto.VariantListDTO) VariantListResponse {
	return VariantListResponse{
		Items:      NewVariantCards(l.Items),
		Pagination: l.Pagination,
		Facets:     l.Facets,
		FacetStats: l.FacetStats,
	}
}
