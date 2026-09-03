package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/delivery/http/request/category_request"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/category_response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/models"
	_ "github.com/stickpro/go-store/internal/storage/base"
	_ "github.com/stickpro/go-store/internal/storage/repository/repository_categories"
)

// getCategoryBySlug is a function get category by slug
//
//	@Summary		Category
//	@Description	Get category by slug
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Category Slug"
//	@Success		200	{object}	response.Result[category_response.CategoryResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/category/{slug}/ [get]
func (h *Handler) getCategoryBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	cat, err := h.services.CategoryService.GetCategoryBySlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "category")
	}
	return c.JSON(response.OkByData(category_response.NewFromModel(cat)))
}

// getCategoryByID is a function get category by id
//
//	@Summary		Category
//	@Description	Get category by id
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Category ID"
//	@Success		200	{object}	response.Result[category_response.CategoryResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/category/id/{id}/ [get]
func (h *Handler) getCategoryByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	cat, err := h.services.CategoryService.GetCategoryByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "category")
	}
	return c.JSON(response.OkByData(category_response.NewFromModel(cat)))
}

// get Categories is a function to Load categories
//
//	@Summary		Get categories
//	@Description	Get categories
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			string	query		category_request.GetCategoryWithPagination	true	"GetCategoriesWithPagination"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[category_response.CategoryResponse]]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/category/ [get]
func (h *Handler) getCategories(c fiber.Ctx) error {
	req := &category_request.GetCategoryWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	cats, err := h.services.CategoryService.GetCategoriesWithPagination(c.Context(), dto.GetDTO{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return h.handleError(err, "category")
	}
	return c.JSON(response.OkByData(category_response.NewPaginated(cats)))
}

// getCategoryTree returns full category hierarchy as a tree
//
//	@Summary		Get category tree
//	@Description	Get all categories as a hierarchical tree structure
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]category_response.CategoryTreeResponse]
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/category/tree [get]
func (h *Handler) getCategoryTree(c fiber.Ctx) error {
	tree, err := h.services.CategoryService.GetCategoryTree(c.Context())
	if err != nil {
		return h.handleError(err, "category")
	}
	return c.JSON(response.OkByData(category_response.NewTree(tree)))
}

// getCategoryProducts returns a paginated, filtered and faceted list of product variants that
// belong to the category (found by slug) or any of its descendants, served from the search index.
// Attribute filters are passed as dynamic query params: "attr.<slug>=v1,v2" for a value set
// (select/text/boolean) and "attr_min.<slug>" / "attr_max.<slug>" for number ranges.
//
//	@Summary		Get category products
//	@Description	Get paginated, filtered and faceted product variants of a category and its subcategories
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string										true	"Category Slug"
//	@Param			string	query		product_request.GetCategoryProductsRequest	true	"Filters, sorting and pagination"
//	@Success		200		{object}	response.Result[product_response.VariantListResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/category/{slug}/products [get]
func (h *Handler) getCategoryProducts(c fiber.Ctx) error {
	req := &product_request.GetCategoryProductsRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	filter := dto.CategoryProductsFilterDTO{
		CategorySlug:  c.Params("slug"),
		Page:          req.Page,
		PageSize:      req.PageSize,
		Sort:          dto.CategoryProductsSort(req.Sort),
		StockStatuses: csvValues(req.StockStatus),
		WithFacets:    req.Facets,
		Attributes:    parseAttributeQueryFilters(c.Queries()),
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

	products, err := h.services.ProductService.SearchVariantsByCategory(c.Context(), filter)
	if err != nil {
		return h.handleError(err, "category products")
	}
	return c.JSON(response.OkByData(product_response.NewVariantList(products)))
}

// csvValues splits a comma-separated query value, trimming blanks.
func csvValues(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseAttributeQueryFilters extracts the dynamic attr.* / attr_min.* / attr_max.* params.
func parseAttributeQueryFilters(queries map[string]string) []dto.AttributeFilterDTO {
	bySlug := make(map[string]*dto.AttributeFilterDTO)
	order := make([]string, 0)
	get := func(slug string) *dto.AttributeFilterDTO {
		if f, ok := bySlug[slug]; ok {
			return f
		}
		f := &dto.AttributeFilterDTO{Slug: slug}
		bySlug[slug] = f
		order = append(order, slug)
		return f
	}

	for key, val := range queries {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		switch {
		case strings.HasPrefix(key, "attr_min."):
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				get(strings.TrimPrefix(key, "attr_min.")).Min = &f
			}
		case strings.HasPrefix(key, "attr_max."):
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				get(strings.TrimPrefix(key, "attr_max.")).Max = &f
			}
		case strings.HasPrefix(key, "attr."):
			slug := strings.TrimPrefix(key, "attr.")
			get(slug).Values = append(get(slug).Values, csvValues(val)...)
		}
	}

	out := make([]dto.AttributeFilterDTO, 0, len(order))
	for _, slug := range order {
		if slug == "" {
			continue
		}
		out = append(out, *bySlug[slug])
	}
	return out
}

// getCategoryFilters returns every filter available for the category listing (price range,
// manufacturers, stock statuses, filterable attributes) with per-option product counts,
// computed over the category and its whole subtree.
//
//	@Summary		Get category filters
//	@Description	Get the full filter set with counts for a category and its subcategories
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string	true	"Category Slug"
//	@Success		200		{object}	response.Result[category_response.CategoryFiltersResponse]
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/category/{slug}/filters [get]
func (h *Handler) getCategoryFilters(c fiber.Ctx) error {
	filters, err := h.services.ProductService.GetCategoryFilters(c.Context(), c.Params("slug"))
	if err != nil {
		return h.handleError(err, "category filters")
	}
	return c.JSON(response.OkByData(category_response.NewFilters(filters)))
}

// getCategoryBreadcrumbs returns the breadcrumb trail (root -> ... -> category) for a category by its slug.
//
//	@Summary		Get category breadcrumbs
//	@Description	Get breadcrumb trail for a category by its slug
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string	true	"Category Slug"
//	@Success		200		{object}	response.Result[[]category_response.BreadcrumbResponse]
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/category/{slug}/breadcrumbs [get]
func (h *Handler) getCategoryBreadcrumbs(c fiber.Ctx) error {
	breadcrumbs, err := h.services.CategoryService.GetBreadcrumbsByCategorySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return h.handleError(err, "category breadcrumbs")
	}
	return c.JSON(response.OkByData(category_response.NewBreadcrumbs(breadcrumbs)))
}

func (h *Handler) initCategoryRoutes(v1 fiber.Router) {
	c := v1.Group("/category")
	c.Get("/", h.getCategories)
	c.Get("/tree", h.getCategoryTree)
	c.Get("/:slug", h.getCategoryBySlug)
	c.Get("/:slug/products", h.getCategoryProducts)
	c.Get("/:slug/filters", h.getCategoryFilters)
	c.Get("/:slug/breadcrumbs", h.getCategoryBreadcrumbs)
	c.Get("/id/:id", h.getCategoryByID)
}
