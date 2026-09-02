package admin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/category_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/category_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/category"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/storage/base"
)

// createCategory is a function create category
//
//	@Summary		Create Category
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
func (h *Handler) createCategory(c fiber.Ctx) error {
	req := &category_request.CreateCategoryRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	dto := category.RequestToCreateDTO(req)
	cat, err := h.services.CategoryService.CreateCategory(c.Context(), dto)
	if err != nil {
		return h.handleError(err, "category")
	}
	return c.JSON(response.OkByData(category_response.NewFromModel(cat)))
}

// updateCategory is a function update category
//
//	@Summary		Update Category
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
func (h *Handler) updateCategory(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &category_request.UpdateCategoryRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if err = req.Validate(id); err != nil {
		return err
	}

	dto := category.RequestToUpdateDTO(req, id)
	cat, err := h.services.CategoryService.UpdateCategory(c.Context(), dto)
	if err != nil {
		return h.handleError(err, "category")
	}

	return c.JSON(response.OkByData(category_response.NewFromModel(cat)))
}

// getCategoryProducts returns a paginated list of product variants that belong to the
// category or any of its descendants, read directly from the database (always consistent).
//
//	@Summary		Get Category Products
//	@Description	Get paginated product variants of a category and its subcategories
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID									true	"Category ID"
//	@Param			string	query		category_request.GetCategoryWithPagination	true	"Pagination params"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[dto.EnrichedVariantDTO]]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/category/:id/products [GET]
//	@Security		BearerAuth
func (h *Handler) getCategoryProducts(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &category_request.GetCategoryWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	products, err := h.services.ProductService.GetEnrichedVariantsByCategoryWithPagination(
		c.Context(),
		id,
		dto.GetDTO{Page: req.Page, PageSize: req.PageSize},
	)
	if err != nil {
		return h.handleError(err, "category products")
	}
	return c.JSON(response.OkByData(products))
}

func (h *Handler) initCategoryRoutes(v1 fiber.Router) {
	c := v1.Group("/category")
	c.Post("/", h.createCategory)
	c.Put("/:id", h.updateCategory)
	c.Get("/:id/products", h.getCategoryProducts)
}
