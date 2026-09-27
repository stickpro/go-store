package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/category_response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
	"github.com/stickpro/go-store/internal/delivery/http/response/resolve_response"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

// resolveSlug resolves a slug to either a category or a product
//
//	@Summary		Resolve slug
//	@Description	Resolve a slug to a category or a product variant. Category is looked up first, then product.
//	@Tags			Resolve
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string	true	"Slug"
//	@Success		200		{object}	response.Result[resolve_response.ResolveResponse]
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@ID				resolveSlug
//	@Router			/v1/resolve/{slug} [get]
func (h *Handler) resolveSlug(c fiber.Ctx) error {
	slug := c.Params("slug")

	cat, err := h.services.CategoryService.GetCategoryBySlug(c.Context(), slug)
	if err == nil {
		return c.JSON(response.OkByData(resolve_response.NewFromCategory(category_response.NewFromModel(cat))))
	}
	if !isSlugNotFoundErr(err) {
		return h.handleError(err, "resolve")
	}

	prd, err := h.services.ProductService.GetProductWithMediaByVariantSlug(c.Context(), slug)
	if err == nil {
		images := h.services.MediaService.Images(prd.Medium, productImageAlt(prd.Variant))
		return c.JSON(response.OkByData(resolve_response.NewFromProduct(product_response.NewFromModelsWithImages(prd.Product, prd.Variant, images))))
	}
	if !isSlugNotFoundErr(err) {
		return h.handleError(err, "resolve")
	}

	return apierror.New().AddError(errors.New("slug not found")).SetHttpCode(fiber.StatusNotFound)
}

// isSlugNotFoundErr reports whether err represents a "no rows" lookup failure.
func isSlugNotFoundErr(err error) bool {
	var notFoundErr *pgerror.NotFoundError
	return errors.Is(err, pgx.ErrNoRows) || errors.As(err, &notFoundErr)
}

func (h *Handler) initResolveRoutes(v1 fiber.Router) {
	v1.Get("/resolve/:slug", h.resolveSlug)
}
