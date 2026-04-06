package admin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen
	_ "github.com/stickpro/go-store/internal/storage/base"
	_ "github.com/stickpro/go-store/internal/storage/repository/repository_product_variants"
	_ "github.com/stickpro/go-store/internal/storage/repository/repository_products"
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
	return c.JSON(response.OkByData(product_response.NewFromProductOnly(prd)))
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
	return c.JSON(response.OkByData(product_response.NewFromProductOnly(prd)))
}

// createProductVariant adds a new variant to an existing product
//
//	@Summary		Create Product Variant
//	@Description	Add a new variant (name/slug/category) to an existing base product
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID									true	"Product ID"
//	@Param			create	body		product_request.CreateProductVariantRequest	true	"Create product variant"
//	@Success		200		{object}	response.Result[product_response.ProductVariantResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/:id/variants [POST]
func (h *Handler) createProductVariant(c fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	req := &product_request.CreateProductVariantRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	var categoryID uuid.NullUUID
	if req.CategoryID != nil {
		categoryID = uuid.NullUUID{UUID: *req.CategoryID, Valid: true}
	}

	d := dto.CreateProductVariantDTO{
		Name:            req.Name,
		Slug:            req.Slug,
		CategoryID:      categoryID,
		Model:           req.Model,
		Description:     req.Description,
		MetaTitle:       req.MetaTitle,
		MetaH1:          req.MetaH1,
		MetaDescription: req.MetaDescription,
		MetaKeyword:     req.MetaKeyword,
		SortOrder:       req.SortOrder,
		IsEnable:        req.IsEnable,
	}

	variant, err := h.services.ProductService.CreateProductVariant(c.Context(), productID, d)
	if err != nil {
		return h.handleError(err, "product variant")
	}
	return c.JSON(response.OkByData(product_response.NewVariantFromModel(variant)))
}

// getProductVariantByID returns a single product variant by its ID
//
//	@Summary		Get Product Variant By ID
//	@Description	Get a single product variant by variant ID
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id			path		uuid.UUID	true	"Product ID"
//	@Param			variant_id	path		uuid.UUID	true	"Variant ID"
//	@Success		200			{object}	response.Result[product_response.ProductVariantResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/product/:id/variants/:variant_id [GET]
//	@Security		BearerAuth
func (h *Handler) getProductVariantByID(c fiber.Ctx) error {
	variantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	variant, err := h.services.ProductService.GetProductVariantByID(c.Context(), variantID)
	if err != nil {
		return h.handleError(err, "product variant")
	}

	return c.JSON(response.OkByData(product_response.NewVariantFromModel(variant)))
}

// getProductVariants returns all variants for a given base product
//
//	@Summary		Get Product Variants
//	@Description	Get all variants (name/slug/category) for an existing base product
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Product ID"
//	@Success		200	{object}	response.Result[[]product_response.ProductVariantResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/product/:id/variants [GET]
//	@Security		BearerAuth
func (h *Handler) getProductVariants(c fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	variants, err := h.services.ProductService.GetProductVariants(c.Context(), productID)
	if err != nil {
		return h.handleError(err, "product variants")
	}
	return c.JSON(response.OkByData(product_response.NewVariantsFromModels(variants)))
}

// updateProductVariant updates an existing product variant
//
//	@Summary		Update Product Variant
//	@Description	Update an existing product variant by its ID
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id			path		uuid.UUID									true	"Product ID"
//	@Param			variant_id	path		uuid.UUID									true	"Variant ID"
//	@Param			update		body		product_request.UpdateProductVariantRequest	true	"Update product variant"
//	@Success		200			{object}	response.Result[product_response.ProductVariantResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/product/:id/variants/:variant_id [PUT]
//	@Security		BearerAuth
func (h *Handler) updateProductVariant(c fiber.Ctx) error {
	variantID, err := uuid.Parse(c.Params("variant_id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	req := &product_request.UpdateProductVariantRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.UpdateProductVariantDTO{
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     req.Description,
		Model:           req.Model,
		MetaTitle:       req.MetaTitle,
		MetaH1:          req.MetaH1,
		MetaDescription: req.MetaDescription,
		MetaKeyword:     req.MetaKeyword,
		Image:           req.Image,
		SortOrder:       req.SortOrder,
		IsEnable:        req.IsEnable,
	}

	if req.CategoryID != nil {
		d.CategoryID = uuid.NullUUID{UUID: *req.CategoryID, Valid: true}
	}

	variant, err := h.services.ProductService.UpdateProductVariant(c.Context(), variantID, d)
	if err != nil {
		return h.handleError(err, "product variant")
	}

	return c.JSON(response.OkByData(product_response.NewVariantFromModel(variant)))
}

// deleteProductVariant deletes a product variant by its ID
//
//	@Summary		Delete Product Variant
//	@Description	Delete a product variant by variant ID
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			id			path		uuid.UUID	true	"Product ID"
//	@Param			variant_id	path		uuid.UUID	true	"Variant ID"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/product/:id/variants/:variant_id [DELETE]
//	@Security		BearerAuth
func (h *Handler) deleteProductVariant(c fiber.Ctx) error {
	variantID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	err = h.services.ProductService.DeleteProductVariant(c.Context(), variantID)
	if err != nil {
		return h.handleError(err, "product variant")
	}

	return c.JSON(response.OkByMessage("Product variant deleted"))
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
		ProductID:         id,
		AttributeValueIDs: req.AttributeValueIDs,
	})
	if err != nil {
		return h.handleError(err, "product")
	}

	return c.JSON(response.OkByMessage("Product attributes synced"))
}

// getProductsWithoutVariants returns a paginated list of products that have no variants
//
//	@Summary		Get products without variants
//	@Description	Get products without variants (admin)
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			string	query		product_request.GetProductWithPagination	true	"GetProductWithPagination"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[repository_products.FindRow]]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/product/list/without-variants [get]
//	@Security		BearerAuth
func (h *Handler) getProductsWithoutVariants(c fiber.Ctx) error {
	req := &product_request.GetProductWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	prds, err := h.services.ProductService.GetProductsWithoutVariants(c.Context(), dto.GetDTO{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(prds))
}

// syncRelatedProducts syncs the list of related variants for a given variant
//
//	@Summary		Sync Related Products
//	@Description	Sync Related Products
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			variant_id	path		uuid.UUID									true	"Variant ID"
//	@Param			update		body		product_request.SyncRelatedProductRequest	true	"Sync related products"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/product/variant/:variant_id/sync-related-products [POST]
func (h *Handler) syncRelatedProducts(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("variant_id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	req := &product_request.SyncRelatedProductRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	err = h.services.ProductService.SyncRelatedProducts(c.Context(), id, req.VariantIDs)
	if err != nil {
		return h.handleError(err, "related products")
	}
	return c.JSON(response.OkByMessage("Related products synced"))
}

// getVariantsWithPagination returns a paginated list of all product variants
//
//	@Summary		Get Variants With Pagination
//	@Description	Get a paginated list of all product variants
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			string	query		product_request.GetVariantWithPagination	true	"Pagination params"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[repository_product_variants.FindRow]]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/product/variants [GET]
//	@Security		BearerAuth
func (h *Handler) getVariantsWithPagination(c fiber.Ctx) error {
	req := &product_request.GetVariantWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	variants, err := h.services.ProductService.GetVariantsWithPagination(c.Context(), dto.GetDTO{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return h.handleError(err, "product variants")
	}
	return c.JSON(response.OkByData(variants))
}

// getVariantCategories returns additional categories for a product variant
//
//	@Summary		Get Variant Categories
//	@Description	Get additional categories for a product variant
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			variant_id	path		uuid.UUID	true	"Variant ID"
//	@Success		200			{object}	response.Result[[]product_response.VariantCategoryResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/product/variant/:variant_id/categories [GET]
//	@Security		BearerAuth
func (h *Handler) getVariantCategories(c fiber.Ctx) error {
	variantID, err := uuid.Parse(c.Params("variant_id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	categories, err := h.services.ProductService.GetVariantCategories(c.Context(), variantID)
	if err != nil {
		return h.handleError(err, "variant categories")
	}
	return c.JSON(response.OkByData(product_response.NewVariantCategoriesFromDTO(categories)))
}

// syncVariantCategories syncs additional categories for a product variant
//
//	@Summary		Sync Variant Categories
//	@Description	Sync additional categories for a product variant
//	@Tags			Product
//	@Accept			json
//	@Produce		json
//	@Param			variant_id	path		uuid.UUID										true	"Variant ID"
//	@Param			update		body		product_request.SyncVariantCategoriesRequest	true	"Sync variant categories"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/product/variant/:variant_id/sync-categories [POST]
//	@Security		BearerAuth
func (h *Handler) syncVariantCategories(c fiber.Ctx) error {
	variantID, err := uuid.Parse(c.Params("variant_id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	req := &product_request.SyncVariantCategoriesRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	err = h.services.ProductService.SyncVariantCategories(c.Context(), dto.SyncVariantCategoriesDTO{
		VariantID:   variantID,
		CategoryIDs: req.CategoryIDs,
	})
	if err != nil {
		return h.handleError(err, "variant categories")
	}
	return c.JSON(response.OkByMessage("Variant categories synced"))
}

func (h *Handler) initProductRoutes(v1 fiber.Router) {
	p := v1.Group("/product")
	p.Post("/", h.createProduct)
	p.Put("/:id", h.updateProduct)
	p.Get("/list/without-variants", h.getProductsWithoutVariants)
	p.Get("/variant/list", h.getVariantsWithPagination)
	p.Get("/:id/variants", h.getProductVariants)
	p.Get("/:id/variants/:variant_id", h.getProductVariantByID)
	p.Post("/:id/variants", h.createProductVariant)
	p.Put("/:id/variants/:variant_id", h.updateProductVariant)
	p.Delete("/:id/variants/:variant_id", h.deleteProductVariant)
	p.Post("/:id/sync-attribute", h.syncProductAttribute)
	p.Post("/variant/:variant_id/sync-related-products", h.syncRelatedProducts)
	p.Get("/variant/:variant_id/categories", h.getVariantCategories)
	p.Post("/variant/:variant_id/sync-categories", h.syncVariantCategories)
}
