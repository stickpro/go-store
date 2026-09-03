package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// searchProducts runs a storefront full-text search over product variants with optional
// filtering by category (slug), price range, manufacturer, stock status and attributes.
// When a query is given and no explicit sort is requested, hits keep their relevance order.
// Attribute filters use the same dynamic params as the category listing:
// "attr.<slug>=v1,v2" and "attr_min.<slug>" / "attr_max.<slug>".
//
//	@Summary		Search products
//	@Description	Full-text product-variant search with category, price and attribute filters
//	@Tags			Search
//	@Accept			json
//	@Produce		json
//	@Param			string	query		product_request.SearchProductsRequest	true	"Query, filters, sorting and pagination"
//	@Success		200		{object}	response.Result[product_response.VariantListResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/search [get]
func (h *Handler) searchProducts(c fiber.Ctx) error {
	req := &product_request.SearchProductsRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	filter := dto.VariantSearchDTO{
		Query:         strings.TrimSpace(req.Query),
		Page:          req.Page,
		PageSize:      req.PageSize,
		Sort:          dto.CategoryProductsSort(req.Sort),
		StockStatuses: csvValues(req.StockStatus),
		WithFacets:    req.Facets,
		Attributes:    parseAttributeQueryFilters(c.Queries()),
	}

	for _, slug := range csvValues(req.Category) {
		cat, err := h.services.CategoryService.GetCategoryBySlug(c.Context(), slug)
		if err != nil {
			return h.handleError(err, "category")
		}
		filter.CategoryIDs = append(filter.CategoryIDs, cat.ID)
	}

	if v := strings.TrimSpace(req.PriceMin); v != "" {
		d, err := decimal.NewFromString(v)
		if err != nil {
			return apierror.New().AddError(fmt.Errorf("invalid price_min: %q", v)).SetHttpCode(fiber.StatusBadRequest)
		}
		filter.PriceMin = &d
	}
	if v := strings.TrimSpace(req.PriceMax); v != "" {
		d, err := decimal.NewFromString(v)
		if err != nil {
			return apierror.New().AddError(fmt.Errorf("invalid price_max: %q", v)).SetHttpCode(fiber.StatusBadRequest)
		}
		filter.PriceMax = &d
	}

	for _, raw := range csvValues(req.ManufacturerID) {
		id, err := uuid.Parse(raw)
		if err != nil {
			return apierror.New().AddError(fmt.Errorf("invalid manufacturer_id: %q", raw)).SetHttpCode(fiber.StatusBadRequest)
		}
		filter.ManufacturerIDs = append(filter.ManufacturerIDs, id)
	}

	res, err := h.services.ProductService.SearchVariants(c.Context(), filter)
	if err != nil {
		return h.handleError(err, "search")
	}
	return c.JSON(response.OkByData(product_response.NewVariantList(res)))
}

func (h *Handler) initSearchRoutes(v1 fiber.Router) {
	v1.Get("/search", h.searchProducts)
}
