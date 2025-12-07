package admin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// createProduct is a function create product
//
//	@Summary		Create Product
//	@Description	Create product
//	@Tags			Product
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
	d := dto.RequestToCreateProductDTO(req)

	prd, err := h.services.ProductService.CreateProduct(c.Context(), d)
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(product_response.NewFromModel(prd)))
}

// updateProduct is a function update product
//
//	@Summary		Update Product
//	@Description	Update product
//	@Tags			Product
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
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	req := &product_request.UpdateProductRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	d := dto.RequestToUpdateProductDTO(req, id)

	prd, err := h.services.ProductService.UpdateProduct(c.Context(), d)
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(product_response.NewFromModel(prd)))
}

// syncProductAttribute is a function sync product attribute
//
//	@Summary		Sync Product Attribute
//	@Description	Sync product attribute
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID									true	"Product ID"
//	@Param			update	body		product_request.SyncProductAttributeRequest	true	"Sync product attribute"
//	@Success		200		{object}	response.Result[string]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/:id/sync-attribute [POST]
func (h *Handler) syncProductAttribute(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	req := &product_request.SyncProductAttributeRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	err = h.services.ProductService.SyncProductAttributes(c.Context(), dto.SyncAttributeProductDTO{
		ProductID:    id,
		AttributeIDs: req.AttributeIDs,
	})
	if err != nil {
		return h.handleError(err, "product")
	}

	return c.JSON(response.OkByMessage("Product attributes synced"))
}

// syncRelatedProduct is a function sync related product
//
//	@Summary		Sync Related Product
//	@Description	Sync Related Product
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID									true	"Product ID"
//	@Param			update	body		product_request.SyncRelatedProductRequest	true	"Sync related product"
//	@Success		200		{object}	response.Result[string]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/:id/sync-related-product [POST]
func (h *Handler) syncRelatedProduct(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	req := &product_request.SyncRelatedProductRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	err = h.services.ProductService.SyncRelatedProduct(c.Context(), id, req.ProductIDs)
	if err != nil {
		return h.handleError(err, "related product")
	}
	return c.JSON(response.OkByMessage("Related product synced"))
}

func (h *Handler) initProductRoutes(v1 fiber.Router) {
	p := v1.Group("/product")
	p.Post("/", h.createProduct)
	p.Put("/:id", h.updateProduct)
	p.Post("/:id/sync-attribute", h.syncProductAttribute)
	p.Post("/:id/sync-related-product", h.syncRelatedProduct)
}
