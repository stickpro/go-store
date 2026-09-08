package admin

import (
	"github.com/gofiber/fiber/v3"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_review_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_review_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools"

	// swag-gen imports
	_ "github.com/stickpro/go-store/internal/storage/base"
	_ "github.com/stickpro/go-store/internal/tools/apierror"
)

// listProductReviews returns reviews across every product, for moderation.
//
//	@Summary		List product reviews
//	@Description	Admin review list. Every filter is optional; deleted reviews are hidden unless with_deleted=true.
//	@Tags			Admin Product Review
//	@Accept			json
//	@Produce		json
//	@Param			request	query		product_review_request.AdminListProductReviewsRequest	true	"Filters + paging"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[product_review_response.AdminProductReviewResponse]]
//	@Failure		400		{object}	apierror.Errors
//	@Router			/v1/admin/product-reviews [get]
//	@Security		BearerAuth
func (h *Handler) listProductReviews(c fiber.Ctx) error {
	req := &product_review_request.AdminListProductReviewsRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	page, err := h.services.ProductReviewService.GetProductReviewsForAdmin(c.Context(), dto.RequestToAdminProductReviewFilter(req))
	if err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByData(product_review_response.NewAdminPaginated(page)))
}

// getProductReview returns one review by id (deleted included).
//
//	@Summary	Get product review
//	@Tags		Admin Product Review
//	@Produce	json
//	@Param		id	path		string	true	"Review ID"
//	@Success	200	{object}	response.Result[product_review_response.AdminProductReviewResponse]
//	@Failure	400	{object}	apierror.Errors
//	@Failure	404	{object}	apierror.Errors
//	@Router		/v1/admin/product-reviews/{id} [get]
//	@Security	BearerAuth
func (h *Handler) getProductReview(c fiber.Ctx) error {
	id, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.services.ProductReviewService.GetProductReviewByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByData(product_review_response.NewAdminFromModel(r)))
}

// updateProductReviewStatus moderates a review (APPROVED publishes it, REJECTED
// hides it, PENDING returns it to the queue).
//
//	@Summary	Update product review status
//	@Tags		Admin Product Review
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string													true	"Review ID"
//	@Param		request	body		product_review_request.UpdateProductReviewStatusRequest	true	"New status"
//	@Success	200		{object}	response.Result[product_review_response.AdminProductReviewResponse]
//	@Failure	400		{object}	apierror.Errors
//	@Failure	404		{object}	apierror.Errors
//	@Router		/v1/admin/product-reviews/{id}/status [patch]
//	@Security	BearerAuth
func (h *Handler) updateProductReviewStatus(c fiber.Ctx) error {
	id, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	req := &product_review_request.UpdateProductReviewStatusRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	r, err := h.services.ProductReviewService.UpdateProductReviewStatus(c.Context(), dto.UpdateProductReviewStatusDTO{
		ID:     id,
		Status: constant.ProductReviewStatus(req.Status),
	})
	if err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByData(product_review_response.NewAdminFromModel(r)))
}

// deleteProductReview soft-deletes a review.
//
//	@Summary	Delete product review
//	@Tags		Admin Product Review
//	@Produce	json
//	@Param		id	path		string	true	"Review ID"
//	@Success	200	{object}	response.Result[any]
//	@Failure	400	{object}	apierror.Errors
//	@Failure	404	{object}	apierror.Errors
//	@Router		/v1/admin/product-reviews/{id} [delete]
//	@Security	BearerAuth
func (h *Handler) deleteProductReview(c fiber.Ctx) error {
	id, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	if err := h.services.ProductReviewService.DeleteProductReview(c.Context(), id); err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByMessage("product review deleted"))
}

// restoreProductReview clears the soft-delete on a review.
//
//	@Summary	Restore product review
//	@Tags		Admin Product Review
//	@Produce	json
//	@Param		id	path		string	true	"Review ID"
//	@Success	200	{object}	response.Result[product_review_response.AdminProductReviewResponse]
//	@Failure	400	{object}	apierror.Errors
//	@Failure	404	{object}	apierror.Errors
//	@Router		/v1/admin/product-reviews/{id}/restore [post]
//	@Security	BearerAuth
func (h *Handler) restoreProductReview(c fiber.Ctx) error {
	id, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.services.ProductReviewService.RestoreProductReview(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByData(product_review_response.NewAdminFromModel(r)))
}

func (h *Handler) initProductReviewRoutes(v1 fiber.Router) {
	r := v1.Group("/admin/product-reviews")
	r.Get("/", h.listProductReviews)
	r.Get("/:id", h.getProductReview)
	r.Patch("/:id/status", h.updateProductReviewStatus)
	r.Delete("/:id", h.deleteProductReview)
	r.Post("/:id/restore", h.restoreProductReview)
}
