package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_review_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/user_response"
	"github.com/stickpro/go-store/internal/dto"

	// swaggo
	_ "github.com/stickpro/go-store/internal/delivery/http/response/product_review_response"
	_ "github.com/stickpro/go-store/internal/tools/apierror"
)

// authUser is a function get user info for auth
//
//	@Summary		Auth user
//	@Description	Auth a user
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[user_response.UserInfoResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		503	{object}	apierror.Errors
//	@Router			/v1/user/info [get]
//	@Security		BearerAuth
func (h *Handler) authUser(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	return c.JSON(response.OkByData(user_response.NewFromModel(user)))
}

// getUserProductReviews is a function get user product reviews
//
//	@Summary		Get user product reviews
//	@Description	Get user product reviews
//	@Tags			Product Review
//	@Accept			json
//	@Produce		json
//	@Param			request	query		product_review_request.GetProductReviewsWithPagination	true	"GetProductReviewsWithPagination"
//	@Success		200		{object}	response.Result[product_review_response.ProductReviewResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
func (h *Handler) getUserProductReviews(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := product_review_request.GetProductReviewsWithPagination{}
	if err := c.Bind().Query(&req); err != nil {
		return err
	}
	d := dto.RequestToGetProductReviewDTO(&req)
	productReviews, err := h.services.ProductReviewService.GetUserProductReviews(c.Context(), d, user.ID)
	if err != nil {
		fmt.Println(err)
		return h.handleError(err, "product reviews")
	}
	return c.JSON(response.OkByData(productReviews))
}

func (h *Handler) initUserRoutes(v1 fiber.Router) {
	u := v1.Group("/user")
	u.Get("/info", h.authUser)
	u.Get("/product-review", h.getUserProductReviews)
}
