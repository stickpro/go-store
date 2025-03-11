package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/category_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/category_response"
	"github.com/stickpro/go-store/internal/service/category"
	"github.com/stickpro/go-store/internal/tools/apierror"
	// swag-gen import
	_ "github.com/stickpro/go-store/internal/storage/base"
	_ "github.com/stickpro/go-store/internal/storage/repository/repository_categories"
)

// getCategoryBySlug is a function get user by slug
//
//	@Summary		Category
//	@Description	Get category by slug
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Category Slug"
//	@Success		200	{object}	response.Result[category_response.CategoryResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/category/:slug/ [get]
func (h Handler) getCategoryBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	cat, err := h.services.CategoryService.GetCategoryBySlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "category")
	}
	return c.JSON(response.OkByData(category_response.NewFromModel(cat)))
}

// getCategoryByID is a function get user by id
//
//	@Summary		Category
//	@Description	Get category by id
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Category ID"
//	@Success		200	{object}	response.Result[category_response.CategoryResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/category/id/:id/ [get]
func (h Handler) getCategoryByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	cat, err := h.services.CategoryService.GetCategoryById(c.Context(), id)
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
//	@Param			string	query		category_request.GetCategoryWithPagination	true	"GetCategoryWithPagination"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[repository_categories.FindRow]]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/category/ [get]
//	@Security		BearerAuth
func (h Handler) getCategories(c fiber.Ctx) error {
	req := &category_request.GetCategoryWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	cats, err := h.services.CategoryService.GetCategoryWithPagination(c.Context(), category.GetDTO{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(cats))
}

func (h *Handler) initCategoryRoutes(v1 fiber.Router) {
	c := v1.Group("/category")
	c.Get("/", h.getCategories)
	c.Get("/:slug", h.getCategoryBySlug)
	c.Get("/id/:id", h.getCategoryByID)
}
