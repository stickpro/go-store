package admin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/stickpro/go-store/internal/delivery/http/request/product_review_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_review_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/storage/base"
)

// listProductReviews returns every review, across all products and statuses,
// for moderation. Filters are optional.
//
//	@Summary		List product reviews
//	@Description	Admin moderation list. Not restricted to APPROVED reviews; `with_deleted=true` also returns soft-deleted rows.
//	@Tags			Admin Product Review
//	@Accept			json
//	@Produce		json
//	@Param			request	query		product_review_request.AdminListProductReviewsRequest	true	"Filters + paging"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[product_review_response.ProductReviewResponse]]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/admin/product-reviews [get]
//	@Security		BearerAuth
func (h *Handler) listProductReviews(c fiber.Ctx) error {
	req := &product_review_request.AdminListProductReviewsRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	d := dto.AdminRequestToListProductReviewsDTO(req)
	reviews, err := h.services.ProductReviewService.GetProductReviewsWithPaginate(c.Context(), d)
	if err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByData(product_review_response.NewPaginated(reviews)))
}

// getProductReview returns one review by its ID, regardless of status.
//
//	@Summary		Get product review
//	@Description	One review by ID, no status filter
//	@Tags			Admin Product Review
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Review ID"
//	@Success		200	{object}	response.Result[product_review_response.ProductReviewResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/admin/product-reviews/{id} [get]
//	@Security		BearerAuth
func (h *Handler) getProductReview(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	review, err := h.services.ProductReviewService.GetProductReviewByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByData(product_review_response.NewFromModel(review)))
}

// updateProductReviewStatus is the moderation decision endpoint: approve, reject
// or send a review back to pending.
//
//	@Summary		Update product review status
//	@Description	Sets the moderation status of a review (PENDING/APPROVED/REJECTED). Only APPROVED reviews are visible on the storefront.
//	@Tags			Admin Product Review
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID												true	"Review ID"
//	@Param			request	body		product_review_request.UpdateProductReviewStatusRequest	true	"New status"
//	@Success		200		{object}	response.Result[product_review_response.ProductReviewResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/admin/product-reviews/{id}/status [patch]
//	@Security		BearerAuth
func (h *Handler) updateProductReviewStatus(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &product_review_request.UpdateProductReviewStatusRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.AdminRequestToUpdateProductReviewStatusDTO(req, id)
	if err := h.services.ProductReviewService.UpdateProductReviewStatus(c.Context(), d); err != nil {
		return h.handleError(err, "product review")
	}

	review, err := h.services.ProductReviewService.GetProductReviewByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByData(product_review_response.NewFromModel(review)))
}

// deleteProductReview soft-deletes a review (keeps the row, hides it from every
// listing). Reversible via the restore endpoint.
//
//	@Summary		Delete product review
//	@Description	Soft delete. The row is kept and can be restored.
//	@Tags			Admin Product Review
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Review ID"
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/admin/product-reviews/{id} [delete]
//	@Security		BearerAuth
func (h *Handler) deleteProductReview(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if err := h.services.ProductReviewService.DeleteProductReview(c.Context(), id); err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByMessage("Product review deleted"))
}

// restoreProductReview clears the soft-delete flag, making the review eligible
// for listing again (subject to its status).
//
//	@Summary		Restore product review
//	@Description	Clears the soft-delete flag set by DELETE.
//	@Tags			Admin Product Review
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Review ID"
//	@Success		200	{object}	response.Result[product_review_response.ProductReviewResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/admin/product-reviews/{id}/restore [post]
//	@Security		BearerAuth
func (h *Handler) restoreProductReview(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	review, err := h.services.ProductReviewService.RestoreProductReview(c.Context(), id)
	if err != nil {
		return h.handleError(err, "product review")
	}

	return c.JSON(response.OkByData(product_review_response.NewFromModel(review)))
}

func (h *Handler) initProductReviewRoutes(v1 fiber.Router) {
	pr := v1.Group("/admin/product-reviews")
	pr.Get("/", h.listProductReviews)
	pr.Get("/:id", h.getProductReview)
	pr.Patch("/:id/status", h.updateProductReviewStatus)
	pr.Delete("/:id", h.deleteProductReview)
	pr.Post("/:id/restore", h.restoreProductReview)
}
