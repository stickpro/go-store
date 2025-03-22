package admin

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
	"github.com/stickpro/go-store/internal/service/product"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

// createProduct is a function create product
//
//	@Summary		Create Product
//	@Description	Create product
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			create	body		product_request.CreateProductRequest	true	"Create product"
//	@Success		200		{object}	response.Result[product_response.ProductResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/ [POST]
func (h *Handler) createProduct(c fiber.Ctx) error {
	req := &product_request.CreateProductRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	dto := product.RequestToCreateDTO(req)

	prd, err := h.services.ProductService.CreateProduct(c.Context(), dto)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}
	return c.JSON(response.OkByData(product_response.NewFromModel(prd)))
}

// updateProduct is a function update product
//
//	@Summary		Update Product
//	@Description	Update product
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID								true	"Product ID"
//	@Param			update	body		product_request.UpdateProductRequest	true	"Update product"
//	@Success		200		{object}	response.Result[product_response.ProductResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/:id [PUT]
func (h *Handler) updateProduct(c fiber.Ctx) error {
	req := &product_request.UpdateProductRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	dto := product.RequestToUpdateDTO(req)

	prd, err := h.services.ProductService.UpdateProduct(c.Context(), dto)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}
	return c.JSON(response.OkByData(product_response.NewFromModel(prd)))
}

func (h *Handler) initProductRoutes(v1 fiber.Router) {
	p := v1.Group("/product")
	p.Post("/", h.createProduct)
	p.Put("/:id", h.updateProduct)
}
