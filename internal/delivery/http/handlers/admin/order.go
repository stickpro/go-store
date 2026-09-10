package admin

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/order_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/order_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/order"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/storage/base"
)

// listOrders returns every order matching the filter, across all accounts and
// guests, newest first.
//
//	@Summary		List orders
//	@Description	Admin order list. Every filter is optional.
//	@Tags			Admin Order
//	@Accept			json
//	@Produce		json
//	@Param			request	query		order_request.AdminListOrdersRequest	true	"Filters + paging"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[order_response.AdminOrderResponse]]
//	@Failure		400		{object}	apierror.Errors
//	@Router			/v1/admin/orders [get]
//	@Security		BearerAuth
func (h *Handler) listOrders(c fiber.Ctx) error {
	req := &order_request.AdminListOrdersRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	filter, err := dto.RequestToAdminOrderFilter(req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	page, err := h.services.OrderService.ListAdmin(c.Context(), filter)
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewAdminPaginated(page)))
}

// getOrder returns one order by its number, regardless of who placed it.
//
//	@Summary		Get order
//	@Description	One order by its number, no ownership check
//	@Tags			Admin Order
//	@Accept			json
//	@Produce		json
//	@Param			number	path		int	true	"Order number"
//	@Success		200		{object}	response.Result[order_response.AdminOrderResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number} [get]
//	@Security		BearerAuth
func (h *Handler) getOrder(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}

	o, err := h.services.OrderService.GetByNumberAdmin(c.Context(), number)
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewAdminFromDTO(o)))
}

// updateOrderStatus drives an order status transition (payment/fulfilment or
// cancellation).
//
//	@Summary		Update order status
//	@Description	Transitions an order to a new status. Invalid transitions from the order's current status return 409.
//	@Tags			Admin Order
//	@Accept			json
//	@Produce		json
//	@Param			number	path		int										true	"Order number"
//	@Param			request	body		order_request.UpdateOrderStatusRequest	true	"New status"
//	@Success		200		{object}	response.Result[order_response.AdminOrderResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number}/status [patch]
//	@Security		BearerAuth
func (h *Handler) updateOrderStatus(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}

	req := &order_request.UpdateOrderStatusRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	admin, ok := c.Locals("user").(*models.User)
	if !ok {
		return apierror.New().AddError(errors.New("undefined user")).SetHttpCode(fiber.StatusUnauthorized)
	}

	o, err := h.services.OrderService.UpdateStatus(c.Context(), number, dto.OrderStatusUpdateDTO{
		Status:        req.Status,
		Actor:         constant.OrderActorAdmin + ":" + admin.ID.String(),
		Comment:       req.Comment,
		PaymentMethod: req.PaymentMethod,
	})
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewAdminFromDTO(o)))
}

// updateOrder edits an order's contact / shipping / payment details.
//
//	@Summary		Update order
//	@Description	Edits an order's contact, shipping address, carrier, payment method and comment. Item lines and prices are not editable. Only orders in status `new` or `pending` can be edited (409 otherwise); editing a `new` (quick) order confirms it into `pending` and requires a shipping address. Send `delivery_method_code` (or `ship_provider` + `ship_tariff_code`) to re-quote the carrier and recompute totals; omit all delivery fields to keep the stored shipping cost.
//	@Tags			Admin Order
//	@Accept			json
//	@Produce		json
//	@Param			number	path		int										true	"Order number"
//	@Param			request	body		order_request.AdminUpdateOrderRequest	true	"Fields to change"
//	@Success		200		{object}	response.Result[order_response.AdminOrderResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number} [patch]
//	@Security		BearerAuth
func (h *Handler) updateOrder(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}

	req := &order_request.AdminUpdateOrderRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	admin, ok := c.Locals("user").(*models.User)
	if !ok {
		return apierror.New().AddError(errors.New("undefined user")).SetHttpCode(fiber.StatusUnauthorized)
	}

	o, err := h.services.OrderService.UpdateDetails(c.Context(), number,
		dto.RequestToOrderDetailsUpdateDTO(req, constant.OrderActorAdmin+":"+admin.ID.String()))
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewAdminFromDTO(o)))
}

// orderError maps order-service errors to HTTP responses. Mirrors the
// customer-facing handler's orderError (internal/delivery/http/handlers/order.go).
func (h *Handler) orderError(err error) error {
	var lineErr *order.LineError
	switch {
	case errors.As(err, &lineErr):
		return apierror.New().AddError(lineErr, apierror.WithField(lineErr.VariantID.String())).
			SetHttpCode(fiber.StatusUnprocessableEntity)
	case errors.Is(err, order.ErrShippingAddressRequired),
		errors.Is(err, order.ErrShippingUnavailable), errors.Is(err, order.ErrShippingMethodUnknown),
		errors.Is(err, shipping.ErrRatesNotSupported):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnprocessableEntity)
	case errors.Is(err, order.ErrInvalidTransition), errors.Is(err, order.ErrDetailsLocked):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusConflict)
	case errors.Is(err, order.ErrNotFound):
		return apierror.New().AddError(errors.New("order not found")).SetHttpCode(fiber.StatusNotFound)
	default:
		return h.handleError(err, "order")
	}
}

func parseOrderNumber(c fiber.Ctx) (int64, error) {
	number, err := strconv.ParseInt(c.Params("number"), 10, 64)
	if err != nil {
		return 0, apierror.New().AddError(errors.New("order number must be an integer")).SetHttpCode(fiber.StatusBadRequest)
	}
	return number, nil
}

// Mounted under /admin, not /orders: the customer-facing order routes
// (internal/delivery/http/handlers/order.go) already claim GET /orders and
// GET /orders/:number on the same fiber.App, and Fiber has no route priority —
// whichever group registers a given method+path first wins it permanently, so
// reusing that path here would leave these handlers unreachable.
func (h *Handler) initOrderRoutes(v1 fiber.Router) {
	o := v1.Group("/admin/orders")
	o.Get("/", h.listOrders)
	o.Get("/:number", h.getOrder)
	o.Patch("/:number", h.updateOrder)
	o.Patch("/:number/status", h.updateOrderStatus)
}
