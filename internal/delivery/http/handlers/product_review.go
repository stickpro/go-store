package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_review_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_review_response"
	"github.com/stickpro/go-store/internal/delivery/middleware"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/tools/apierror"
)

// createProductReview
//
//	@Summary		Create Product Review
//	@Tags			Product Review
//	@Description	Create Product Review
//	@Accept			json
//	@Produce		json
//	@Param			request	body		product_review_request.CreateProductReviewRequest	true	"Create Product Review Request"
//	@Success		200		{object}	response.Result[product_review_response.ProductReviewResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		403		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Security		BearerAuth
//	@Router			/api/v1/product_review/ [post]
func (h *Handler) createProductReview(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := product_review_request.CreateProductReviewRequest{}

	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	d := dto.RequestToCreateProductReviewDTO(&req, usr.ID)
	productReview, err := h.services.ProductReviewService.CreateProductReview(c.Context(), d)
	if err != nil {
		return h.handleError(err, "product")
	}
	return c.JSON(response.OkByData(product_review_response.NewFromModel(productReview)))
}

// getProductReviewsByProductID
//
//	@Summary		Get Product Reviews By Product ID
//	@Tags			Product Review
//	@Description	Get Product Reviews By Product ID
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID												true	"Product ID"
//	@Param			request	query		product_review_request.GetProductReviewsWithPagination	true	"GetProductReviewsWithPagination"
//	@Success		200		{object}	response.Result[product_review_response.ProductReviewResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/api/v1/product_review/by-product/{id} [post]
func (h *Handler) getProductReviewsByProductID(c fiber.Ctx) error {
	id, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	req := product_review_request.GetProductReviewsWithPagination{}
	if err := c.Bind().Query(&req); err != nil {
		return err
	}
	d := dto.RequestToGetProductReviewDTO(&req)
	productReviews, err := h.services.ProductReviewService.GetProductReviewsByProductID(c.Context(), d, &id)
	if err != nil {
		return h.handleError(err, "product reviews")
	}
	return c.JSON(response.OkByData(productReviews))
}

func (h *Handler) initProductReviewRoutes(v1 fiber.Router) {
	pr := v1.Group("/product-review")
	pr.Post("/", h.createProductReview, middleware.AuthMiddleware(h.services.AuthService))
	pr.Get("/by-product/:id", h.getProductReviewsByProductID)
}
