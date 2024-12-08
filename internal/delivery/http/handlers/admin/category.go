package admin

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/category_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/category_response"
	"github.com/stickpro/go-store/internal/service/category"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

// createCategory is a function create category
//
//	@Summary		Category
//	@Description	Create category
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			create	body		category_request.CreateCategoryRequest	true	"Create category"
//	@Success		200		{object}	response.Result[category_response.CategoryResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/category/ [POST]
func (h Handler) createCategory(c fiber.Ctx) error {
	req := &category_request.CreateCategoryRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	dto := category.RequestToCreateDTO(req)

	cat, err := h.services.CategoryService.CreateCategory(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(category_response.NewFromModel(cat)))
}

// updateCategory is a function update category
//
//	@Summary		Category
//	@Description	Update category
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID								true	"Category ID"
//	@Param			update	body		category_request.UpdateCategoryRequest	true	"Update category"
//	@Success		200		{object}	response.Result[category_response.CategoryResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/category/:id [PUT]
func (h Handler) updateCategory(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &category_request.UpdateCategoryRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	dto := category.RequestToUpdateDTO(req, id)
	cat, err := h.services.CategoryService.UpdateCategory(c.Context(), dto)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}

	return c.JSON(response.OkByData(category_response.NewFromModel(cat)))
}

func (h *Handler) initCategoryRoutes(v1 fiber.Router) {
	c := v1.Group("/category")
	c.Post("/", h.createCategory)
	c.Put("/:id", h.updateCategory)
}
