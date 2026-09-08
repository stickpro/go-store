package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/sitemap_response"
)

// getSitemapProducts returns every enabled product-variant page as slug +
// updated_at, for the frontend sitemap generator.
//
//	@Summary		Sitemap: products
//	@Description	Flat feed of enabled product-variant pages (slug + updated_at). Unpaginated.
//	@Tags			Sitemap
//	@Produce		json
//	@Success		200	{object}	response.Result[[]sitemap_response.SitemapEntry]
//	@Router			/v1/sitemap/products [get]
func (h *Handler) getSitemapProducts(c fiber.Ctx) error {
	items, err := h.services.ProductService.GetSitemapEntries(c.Context())
	if err != nil {
		return h.handleError(err, "sitemap")
	}
	return c.JSON(response.OkByData(sitemap_response.NewFromDTOs(items)))
}

// getSitemapCategories returns every enabled category page as slug + updated_at.
//
//	@Summary		Sitemap: categories
//	@Description	Flat feed of enabled category pages (slug + updated_at). Unpaginated.
//	@Tags			Sitemap
//	@Produce		json
//	@Success		200	{object}	response.Result[[]sitemap_response.SitemapEntry]
//	@Router			/v1/sitemap/categories [get]
func (h *Handler) getSitemapCategories(c fiber.Ctx) error {
	items, err := h.services.CategoryService.GetSitemapEntries(c.Context())
	if err != nil {
		return h.handleError(err, "sitemap")
	}
	return c.JSON(response.OkByData(sitemap_response.NewFromDTOs(items)))
}

func (h *Handler) initSitemapRoutes(v1 fiber.Router) {
	s := v1.Group("/sitemap")
	s.Get("/products", h.getSitemapProducts)
	s.Get("/categories", h.getSitemapCategories)
}
