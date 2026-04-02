package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/models"
	_ "github.com/stickpro/go-store/internal/storage/base"
	_ "github.com/stickpro/go-store/internal/storage/repository/repository_products"
)

// getProductBySlug returns a product with media by variant slug
//
//	@Summary		Product
//	@Description	Get product by slug
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Product Slug"
//	@Success		200	{object}	response.Result[product_response.ProductWithMediumResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/product/{slug}/ [get]
func (h *Handler) getProductBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	prd, err := h.services.ProductService.GetProductWithMediaByVariantSlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "product")
	}

	return c.JSON(response.OkByData(product_response.NewFromModelsWithMedium(prd.Product, prd.Variant, prd.Medium)))
}

// getProductByID returns a product by its base product ID
//
//	@Summary		Product
//	@Description	Get product by id
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Product ID"
//	@Success		200	{object}	response.Result[product_response.ProductResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/product/id/{id}/ [get]
func (h *Handler) getProductByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	prd, err := h.services.ProductService.GetProductByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(product_response.NewFromProductOnly(prd)))
}

// getProducts returns a paginated list of products
//
//	@Summary		Get products
//	@Description	Get products
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			string	query		product_request.GetProductWithPagination	true	"GetProductWithPagination"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[repository_products.FindRow]]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/product/ [get]
//	@Security		BearerAuth
func (h *Handler) getProducts(c fiber.Ctx) error {
	req := &product_request.GetProductWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	prds, err := h.services.ProductService.GetProductWithPagination(c.Context(), dto.GetDTO{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(prds))
}

// getProductWithMediaByID returns a product with media by its base product ID
//
//	@Summary		Get product with media by ID
//	@Description	Get product with media by ID
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Product ID"
//	@Success		200	{object}	response.Result[product_response.ProductWithMediumResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/product/id/{id}/with-media [get]
func (h *Handler) getProductWithMediaByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	prd, err := h.services.ProductService.GetProductWithMediaByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(product_response.NewFromModelsWithMedium(prd.Product, prd.Variant, prd.Medium)))
}

// findProduct searches for a product by name via search index
//
//	@Summary		Find product
//	@Description	Find product by name
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			product	query		string	true	"Product name"
//	@Success		200		{object}	response.Result[[]models.Product]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/find [get]
func (h *Handler) findProduct(c fiber.Ctx) error {
	product := c.Query("product")
	if product == "" {
		return apierror.New().AddError(fmt.Errorf("product is requered")).SetHttpCode(fiber.StatusBadRequest)
	}
	products, err := h.services.SearchService.Search(constant.ProductVariantsIndex, product, 10, 0)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(products.Hits))
}

// getProductAttributes returns all attribute groups with values for a product by variant slug
//
//	@Summary		Get product attributes
//	@Description	Get product attributes by variant slug
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string	true	"Product Slug"
//	@Success		200		{object}	response.Result[product_response.ProductAttributeResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/{slug}/attributes [get]
func (h *Handler) getProductAttributes(c fiber.Ctx) error {
	slug := c.Params("slug")
	prdAttr, err := h.services.ProductService.GetProductAttributesBySlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(product_response.NewFromAttributeWithAttributeGroups(prdAttr)))
}

// getProductBreadcrumbs returns the breadcrumb trail for a product by variant slug
//
//	@Summary		Get product breadcrumbs
//	@Description	Get breadcrumb trail for a product by its slug
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string	true	"Product Slug"
//	@Success		200		{object}	response.Result[[]dto.BreadcrumbDTO]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/{slug}/breadcrumbs [get]
func (h *Handler) getProductBreadcrumbs(c fiber.Ctx) error {
	slug := c.Params("slug")
	breadcrumbs, err := h.services.CategoryService.GetBreadcrumbsByProductSlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "product breadcrumbs")
	}
	return c.JSON(response.OkByData(breadcrumbs))
}

// getRelatedProducts returns related products for a variant by variant ID
//
//	@Summary		Get related products
//	@Description	Get related products by variant ID
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Variant ID"
//	@Success		200	{object}	response.Result[[]models.ShortProduct]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/product/variant/id/:id/related-products [get]
func (h *Handler) getRelatedProducts(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	prd, err := h.services.ProductService.GetRelatedProducts(c.Context(), id)
	if err != nil {
		return h.handleError(err, "related products")
	}
	return c.JSON(response.OkByData(prd))
}

// getRelatedProductsBatch returns related products for multiple variants in one request
//
//	@Summary		Get related products batch
//	@Description	Get related products for multiple variants in one request
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			get		body		product_request.GetRelatedProductsBatchRequest	true	"List of variant IDs"
//	@Success		200		{object}	response.Result[map[string][]models.ShortProduct]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/variant/related-products/batch [post]
func (h *Handler) getRelatedProductsBatch(c fiber.Ctx) error {
	req := &product_request.GetRelatedProductsBatchRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	result, err := h.services.ProductService.GetRelatedProductsBatch(c.Context(), req.VariantIDs)
	if err != nil {
		return h.handleError(err, "related products")
	}
	return c.JSON(response.OkByData(result))
}

// getRelatedProductsBySlug returns related products for a variant by variant slug
//
//	@Summary		Get related products by slug
//	@Description	Get related products by variant slug
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string	true	"Variant slug"
//	@Success		200		{object}	response.Result[[]models.ShortProduct]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/:slug/related-products [get]
func (h *Handler) getRelatedProductsBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	prd, err := h.services.ProductService.GetRelatedProductsBySlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "related products")
	}
	return c.JSON(response.OkByData(prd))
}

// getProductAttributesByID returns all attribute groups with values for a product by product ID
//
//	@Summary		Get product attributes
//	@Description	Get product attributes by product ID
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Product ID"
//	@Success		200	{object}	response.Result[product_response.ProductAttributeResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/product/{id}/attributes [get]
func (h *Handler) getProductAttributesByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	prdAttr, err := h.services.ProductService.GetProductAttributesByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(product_response.NewFromAttributeWithAttributeGroups(prdAttr)))
}

func (h *Handler) initProductRoutes(v1 fiber.Router) {
	p := v1.Group("/product")
	p.Get("/", h.getProducts)
	p.Get("/find", h.findProduct)
	// by slug
	p.Get("/:slug", h.getProductBySlug)
	p.Get("/:slug/attributes", h.getProductAttributes)
	p.Get("/:slug/breadcrumbs", h.getProductBreadcrumbs)
	p.Get("/:slug/related-products", h.getRelatedProductsBySlug)
	p.Get("/:slug/reviews", h.getProductReviewsBySlug)
	// by product id
	p.Get("/id/:id", h.getProductByID)
	p.Get("/id/:id/with-media", h.getProductWithMediaByID)
	p.Get("/id/:id/attributes", h.getProductAttributesByID)
	p.Get("/id/:id/reviews", h.getProductReviewsByProductID)
	// by variant id
	p.Get("/variant/id/:id/related-products", h.getRelatedProducts)
	p.Post("/variant/related-products/batch", h.getRelatedProductsBatch)
}
