package admin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/request/category_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/category_response"
	"github.com/stickpro/go-store/internal/service/category"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// createCategory is a function create category
//
//	@Summary		Category
//	@Description	Create category
//	@Tags			Auth
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
	dto := category.RequestToDTO(req)

	cat, err := h.services.CategoryService.CreateCategory(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(category_response.NewFromModel(cat)))

}

func (h *Handler) initCategoryRoutes(v1 fiber.Router) {
	c := v1.Group("/category")
	c.Post("/", h.createCategory)
}
