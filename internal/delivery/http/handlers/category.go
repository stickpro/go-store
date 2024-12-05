package handlers

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/category_response"
	"github.com/stickpro/go-store/internal/tools/apierror"
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
	category, err := h.services.CategoryService.GetCategoryBySlug(c.Context(), slug)
	if err != nil {
		preparedErr := apierror.New().AddError(err)
		if errors.Is(err, pgx.ErrNoRows) {
			return preparedErr.SetHttpCode(fiber.StatusNotFound)
		}
		return preparedErr.SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(category_response.NewFromModel(category)))
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
//	@Router			/v1/category/:id/ [get]
func (h Handler) getCategoryByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	category, err := h.services.CategoryService.GetCategoryById(c.Context(), id)
	if err != nil {
		preparedErr := apierror.New().AddError(err)
		if errors.Is(err, pgx.ErrNoRows) {
			return preparedErr.SetHttpCode(fiber.StatusNotFound)
		}
		return preparedErr.SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(category_response.NewFromModel(category)))
}

func (h *Handler) initCategoryRoutes(v1 fiber.Router) {
	c := v1.Group("/category")
	c.Get("/:slug", h.getCategoryBySlug)
	c.Get("/:id", h.getCategoryByID)
}
