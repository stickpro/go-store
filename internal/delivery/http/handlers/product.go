package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
	"github.com/stickpro/go-store/internal/service/product"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/models"
	_ "github.com/stickpro/go-store/internal/storage/base"
	_ "github.com/stickpro/go-store/internal/storage/repository/repository_products"
)

// getProductBySlug is a function get user by slug
//
//	@Summary		Product
//	@Description	Get product by slug
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Product Slug"
//	@Success		200	{object}	response.Result[product_response.ProductResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/product/:slug/ [get]
func (h Handler) getProductBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	prd, err := h.services.ProductService.GetProductBySlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "product")
	}

	return c.JSON(response.OkByData(product_response.NewFromModel(prd)))
}

// getProductByID is a function get user by id
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
//	@Router			/v1/product/id/:id/ [get]
func (h Handler) getProductByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	prd, err := h.services.ProductService.GetProductById(c.Context(), id)
	if err != nil {
		return h.handleError(err, "category")
	}
	return c.JSON(response.OkByData(product_response.NewFromModel(prd)))
}

// getProducts is a function to Load products
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
func (h Handler) getProducts(c fiber.Ctx) error {
	req := &product_request.GetProductWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	prds, err := h.services.ProductService.GetProductWithPagination(c.Context(), product.GetDTO{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(prds))
}

// getProductWithMediumByID is a function to Load product with medium by ID
//
//	@Summary		Get product with medium by ID
//	@Description	Get product with medium by ID
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Product ID"
//	@Success		200	{object}	response.Result[product_response.ProductWithMediumResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/product/id/:id/with-medium [get]
func (h Handler) getProductWithMediumByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	prd, err := h.services.ProductService.GetProductWithMediumByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(product_response.NewFromModelWithMedium(prd.Product, prd.Medium)))
}

// findProduct is a function find product by name
//
//	@Summary		Find product
//	@Description	Find product by name
//	@Tags			Geo
//	@Accept			json
//	@Produce		json
//	@Param			product	query		string	true	"Product name"
//	@Success		200		{object}	response.Result[[]models.Product]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/find [get]
func (h *Handler) findProduct(c fiber.Ctx) error {
	city := c.Query("product")
	if city == "" {
		return apierror.New().AddError(fmt.Errorf("product is requered")).SetHttpCode(fiber.StatusBadRequest)

	}
	location, err := h.services.SearchService.Search(constant.ProductsIndex, city, 10, 0)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(location.Hits))
}

func (h *Handler) initProductRoutes(v1 fiber.Router) {
	p := v1.Group("/product")
	p.Get("/", h.getProducts)
	p.Get("/find", h.findProduct)
	p.Get("/:slug", h.getProductBySlug)
	p.Get("/id/:id", h.getProductByID)
	p.Get("/id/:id/with-medium", h.getProductWithMediumByID)
}
