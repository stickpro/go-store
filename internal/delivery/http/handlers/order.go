package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/stickpro/go-store/internal/delivery/http/request/order_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/order_response"
	"github.com/stickpro/go-store/internal/delivery/middleware"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/order"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/storage/base"
	_ "github.com/stickpro/go-store/internal/tools/apierror"
)

// createOrder turns the caller's cart into an order.
//
//	@Summary		Create order
//	@Description	Checkout: converts the cart (session or account) into an order, decrements stock and clears the cart. Send an `Idempotency-Key` header (UUID) to make retries safe.
//	@Tags			Order
//	@Accept			json
//	@Produce		json
//	@Param			Idempotency-Key	header		string								false	"Idempotency key (UUID)"
//	@Param			request			body		order_request.CreateOrderRequest	true	"Checkout data"
//	@Success		200				{object}	response.Result[order_response.OrderResponse]
//	@Failure		400				{object}	apierror.Errors
//	@Failure		409				{object}	apierror.Errors
//	@Failure		422				{object}	apierror.Errors
//	@Failure		500				{object}	apierror.Errors
//	@Router			/v1/orders [post]
func (h *Handler) createOrder(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}

	req := &order_request.CreateOrderRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	// Optional auth middleware populates the user when a valid token is present.
	var user *models.User
	if u, aErr := loadAuthUser(c); aErr == nil {
		user = u
	}

	d, err := dto.RequestToCreateOrderDTO(req, owner, user)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnprocessableEntity)
	}
	if key := c.Get("Idempotency-Key"); key != "" {
		d.IdempotencyKey = &key
	}

	created, err := h.services.OrderService.CreateOrder(c.Context(), d)
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewFromDTO(created)))
}

// previewCheckout returns the server-computed cart total for a delivery choice.
//
//	@Summary		Checkout preview
//	@Description	Server-computed money breakdown (subtotal + shipping + grand total) for the caller's cart under a delivery choice. Send `delivery_method_code` from GET /v1/delivery/methods plus `ship_point_code` (pickup) or `ship_postcode` (courier). Nothing is persisted; the frontend renders `grand_total` from here instead of adding shipping to the cart total itself.
//	@Tags			Order
//	@Accept			json
//	@Produce		json
//	@Param			request	body		order_request.CheckoutPreviewRequest	true	"Delivery choice"
//	@Success		200		{object}	response.Result[order_response.CheckoutPreviewResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/orders/preview [post]
func (h *Handler) previewCheckout(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}

	req := &order_request.CheckoutPreviewRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	var user *models.User
	if u, aErr := loadAuthUser(c); aErr == nil {
		user = u
	}

	result, err := h.services.OrderService.PreviewCheckout(c.Context(), dto.RequestToCheckoutPreviewDTO(req, owner, user))
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewCheckoutPreviewFromDTO(result)))
}

// listOrders returns the authenticated account's orders, newest first.
//
//	@Summary		List orders
//	@Description	Order history for the authenticated account
//	@Tags			Order
//	@Accept			json
//	@Produce		json
//	@Param			request	query		order_request.ListOrdersRequest	true	"Paging"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[order_response.OrderResponse]]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Router			/v1/orders [get]
//	@Security		BearerAuth
func (h *Handler) listOrders(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &order_request.ListOrdersRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	page, err := h.services.OrderService.ListForUser(c.Context(), user.ID, dto.RequestToListOrdersDTO(req))
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewPaginated(page)))
}

// getOrder returns one of the authenticated account's orders by its number.
//
//	@Summary		Get order
//	@Description	One order by its number, scoped to the authenticated account
//	@Tags			Order
//	@Accept			json
//	@Produce		json
//	@Param			number	path		int	true	"Order number"
//	@Success		200		{object}	response.Result[order_response.OrderResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/orders/{number} [get]
//	@Security		BearerAuth
func (h *Handler) getOrder(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	number, err := strconv.ParseInt(c.Params("number"), 10, 64)
	if err != nil {
		return apierror.New().AddError(errors.New("order number must be an integer")).SetHttpCode(fiber.StatusBadRequest)
	}

	o, err := h.services.OrderService.GetByNumber(c.Context(), user.ID, number)
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewFromDTO(o)))
}

// orderError maps order-service errors to HTTP responses.
func (h *Handler) orderError(err error) error {
	var lineErr *order.LineError
	switch {
	case errors.As(err, &lineErr):
		return apierror.New().AddError(lineErr, apierror.WithField(lineErr.VariantID.String())).
			SetHttpCode(fiber.StatusUnprocessableEntity)
	case errors.Is(err, order.ErrCartEmpty), errors.Is(err, dto.ErrEmailRequired),
		errors.Is(err, order.ErrShippingUnavailable), errors.Is(err, order.ErrShippingMethodUnknown),
		errors.Is(err, shipping.ErrRatesNotSupported):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnprocessableEntity)
	case errors.Is(err, order.ErrPriceChanged), errors.Is(err, order.ErrInvalidTransition):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusConflict)
	case errors.Is(err, order.ErrNotFound), errors.Is(err, order.ErrForbidden):
		return apierror.New().AddError(errors.New("order not found")).SetHttpCode(fiber.StatusNotFound)
	default:
		return h.handleError(err, "order")
	}
}

func (h *Handler) initOrderRoutes(v1 fiber.Router) {
	o := v1.Group("/orders")
	o.Post("/", middleware.OptionalAuthMiddleware(h.services.AuthService), h.createOrder)
	o.Post("/preview", middleware.OptionalAuthMiddleware(h.services.AuthService), h.previewCheckout)
	o.Get("/", middleware.AuthMiddleware(h.services.AuthService), h.listOrders)
	o.Get("/:number", middleware.AuthMiddleware(h.services.AuthService), h.getOrder)
}
