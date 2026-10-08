package admin

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/order_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/order_response"
	"github.com/stickpro/go-store/internal/delivery/http/response/payment_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/order"
	"github.com/stickpro/go-store/internal/service/payment"
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

// previewOrderItems computes an item edit without saving it.
//
//	@Summary		Preview order item edit
//	@Description	Runs the item edit exactly as PUT would (stock, prices, totals, paid-order limit) and rolls it back. `lines` is the full target line set. Use the returned `order.version` and `order.grand_total` as `expected_version` / `expected_grand_total` for the PUT.
//	@Tags			Admin Order
//	@Accept			json
//	@Produce		json
//	@Param			number	path		int											true	"Order number"
//	@Param			request	body		order_request.AdminPreviewOrderItemsRequest	true	"Target line set"
//	@Success		200		{object}	response.Result[order_response.AdminOrderItemsEditResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number}/items/preview [post]
//	@Security		BearerAuth
func (h *Handler) previewOrderItems(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}

	req := &order_request.AdminPreviewOrderItemsRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	admin, ok := c.Locals("user").(*models.User)
	if !ok {
		return apierror.New().AddError(errors.New("undefined user")).SetHttpCode(fiber.StatusUnauthorized)
	}

	res, err := h.services.OrderService.EditItems(c.Context(), number,
		dto.RequestToOrderItemsPreviewDTO(req, constant.OrderActorAdmin+":"+admin.ID.String()))
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewItemsEditFromDTO(res)))
}

// updateOrderItems replaces the order's lines.
//
//	@Summary		Edit order items
//	@Description	Replaces the order's lines with `lines` (the full target set: a line left out is removed). Kept lines keep the price they were sold at; added lines are priced from the catalogue; stock moves by the difference. Allowed while the order is `new`, `pending`, `paid` or `processing` (409 otherwise). For a paid order the total may only go down: `refund_due` in the response is the overpayment to return via POST /refund (409 if the edit would need an extra payment). For an unpaid order whose total changed, the open payment link is voided — the customer starts a new payment for the new total. 409 if `expected_version` / `expected_grand_total` no longer match (the order changed since the preview).
//	@Tags			Admin Order
//	@Accept			json
//	@Produce		json
//	@Param			number	path		int											true	"Order number"
//	@Param			request	body		order_request.AdminUpdateOrderItemsRequest	true	"Target line set and the preview's version/total"
//	@Success		200		{object}	response.Result[order_response.AdminOrderItemsEditResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number}/items [put]
//	@Security		BearerAuth
func (h *Handler) updateOrderItems(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}

	req := &order_request.AdminUpdateOrderItemsRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	admin, ok := c.Locals("user").(*models.User)
	if !ok {
		return apierror.New().AddError(errors.New("undefined user")).SetHttpCode(fiber.StatusUnauthorized)
	}

	res, err := h.services.OrderService.EditItems(c.Context(), number,
		dto.RequestToOrderItemsEditDTO(req, constant.OrderActorAdmin+":"+admin.ID.String()))
	if err != nil {
		return h.orderError(err)
	}

	// An unpaid order's open PaymentURL was issued for the old total. Voiding
	// it is best-effort: if it fails and the customer pays the old link, the
	// webhook refuses to mark the order paid for a mismatched amount.
	unpaid := res.Order.PaymentStatus == constant.PaymentUnpaid.String() || res.Order.PaymentStatus == constant.PaymentFailed.String()
	if unpaid && !res.PreviousGrandTotal.Equal(res.Order.GrandTotal) {
		if err := h.services.PaymentService.VoidPending(c.Context(), res.Order.ID); err != nil {
			h.logger.Errorw("admin: void payment link after item edit", "order", res.Order.Number, "error", err)
		}
	}

	return c.JSON(response.OkByData(order_response.NewItemsEditFromDTO(res)))
}

// listOrderEdits returns the order's item-edit audit trail.
//
//	@Summary		List order item edits
//	@Description	Every item edit of the order, newest first, with the full line set and grand total before and after.
//	@Tags			Admin Order
//	@Produce		json
//	@Param			number	path		int	true	"Order number"
//	@Success		200		{object}	response.Result[[]order_response.OrderEditResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number}/edits [get]
//	@Security		BearerAuth
func (h *Handler) listOrderEdits(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}

	edits, err := h.services.OrderService.ListEdits(c.Context(), number)
	if err != nil {
		return h.orderError(err)
	}

	return c.JSON(response.OkByData(order_response.NewEditList(edits)))
}

// refundOrder refunds the overpayment (or, with full, everything) from the
// order's captured payment at the provider, then brings the order in line: a
// full refund moves it to "refunded" (with restock) when its status allows it,
// a partial one only updates payment_status / refunded_total.
//
//	@Summary		Refund order
//	@Description	Refunds the order's captured payment at the acquirer. The amount is computed, never sent: by default it is the overpayment — what the customer paid above the order's current `grand_total` (e.g. after items were removed; see `refund_due` of the item edit). That leaves the order's status and stock alone and sets `payment_status` to `partially_refunded`. With `full: true` everything not yet refunded is returned and the order moves to `refunded` (with restock) when its status allows it. The `Idempotency-Key` header (UUID) is required: retrying with the same key returns the original refund instead of refunding again — on a timeout, retry with the SAME key. 409 if nothing is owed, nothing captured is left, or a refund is still pending at the acquirer; 422 if the key was used for another order's refund or the acquirer declined.
//	@Tags			Admin Order
//	@Accept			json
//	@Produce		json
//	@Param			number			path		int										true	"Order number"
//	@Param			Idempotency-Key	header		string									true	"Idempotency key (UUID)"
//	@Param			request			body		order_request.AdminRefundOrderRequest	false	"Refund mode and reason"
//	@Success		200				{object}	response.Result[order_response.AdminOrderResponse]
//	@Failure		400				{object}	apierror.Errors
//	@Failure		404				{object}	apierror.Errors
//	@Failure		409				{object}	apierror.Errors
//	@Failure		422				{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number}/refund [post]
//	@Security		BearerAuth
func (h *Handler) refundOrder(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}

	key := c.Get("Idempotency-Key")
	if _, err := uuid.Parse(key); err != nil {
		return apierror.New().AddError(errors.New("Idempotency-Key header must be a UUID")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &order_request.AdminRefundOrderRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	admin, ok := c.Locals("user").(*models.User)
	if !ok {
		return apierror.New().AddError(errors.New("undefined user")).SetHttpCode(fiber.StatusUnauthorized)
	}
	actor := constant.OrderActorAdmin + ":" + admin.ID.String()

	o, err := h.services.OrderService.GetByNumberAdmin(c.Context(), number)
	if err != nil {
		return h.orderError(err)
	}

	res, err := h.services.PaymentService.Refund(c.Context(), dto.RefundPaymentDTO{
		OrderID:        o.ID,
		Full:           req.Full,
		IdempotencyKey: key,
		Reason:         req.Reason,
		Actor:          actor,
	})
	if err != nil {
		return h.paymentError(err)
	}

	// The money has moved at this point: a failure to update the order is
	// logged and the order returned as it is, rather than reporting the
	// refund itself as failed.
	updated, err := h.services.OrderService.RecordRefund(c.Context(), o.ID, res.Payment.RefundedAmount,
		payment.Status(res.Payment.Status) == payment.StatusRefunded, actor)
	if err != nil {
		h.logger.Errorw("admin: refund done but order not updated", "order", o.Number, "refund", res.Refund.ID, "error", err)
		updated, err = h.services.OrderService.GetByNumberAdmin(c.Context(), number)
		if err != nil {
			return h.orderError(err)
		}
	}

	return c.JSON(response.OkByData(order_response.NewAdminFromDTO(updated)))
}

// listOrderRefunds returns every refund attempt made for the order.
//
//	@Summary		List order refunds
//	@Description	Every refund attempt for the order, newest first, including pending (outcome not yet known at the acquirer) and failed ones.
//	@Tags			Admin Order
//	@Produce		json
//	@Param			number	path		int	true	"Order number"
//	@Success		200		{object}	response.Result[[]payment_response.RefundResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number}/refunds [get]
//	@Security		BearerAuth
func (h *Handler) listOrderRefunds(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}

	o, err := h.services.OrderService.GetByNumberAdmin(c.Context(), number)
	if err != nil {
		return h.orderError(err)
	}

	refunds, err := h.services.PaymentService.ListRefunds(c.Context(), o.ID)
	if err != nil {
		return h.paymentError(err)
	}

	return c.JSON(response.OkByData(payment_response.NewRefundList(refunds)))
}

// paymentError maps payment-service errors to HTTP responses.
func (h *Handler) paymentError(err error) error {
	switch {
	case errors.Is(err, payment.ErrProviderNotFound):
		return apierror.New().AddError(errors.New("unknown payment provider")).SetHttpCode(fiber.StatusUnprocessableEntity)
	case errors.Is(err, payment.ErrNotFound):
		return apierror.New().AddError(errors.New("payment not found")).SetHttpCode(fiber.StatusNotFound)
	case errors.Is(err, payment.ErrNotConfirmed):
		return apierror.New().AddError(errors.New("order has no captured payment left to refund")).SetHttpCode(fiber.StatusConflict)
	case errors.Is(err, payment.ErrNothingToRefund):
		return apierror.New().AddError(errors.New("nothing to refund: the customer has paid no more than the order total; send full=true to refund the whole order")).SetHttpCode(fiber.StatusConflict)
	case errors.Is(err, payment.ErrRefundPending):
		h.logger.Warnw("admin: refund pending", "error", err)
		return apierror.New().AddError(errors.New("refund is pending at the acquirer; retry later with the same Idempotency-Key, never a new one")).SetHttpCode(fiber.StatusConflict)
	case errors.Is(err, payment.ErrIdempotencyConflict),
		errors.Is(err, payment.ErrRefundRejected):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnprocessableEntity)
	default:
		return h.handleError(err, "payment")
	}
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
	case errors.Is(err, order.ErrInvalidTransition), errors.Is(err, order.ErrDetailsLocked),
		errors.Is(err, order.ErrItemsLocked), errors.Is(err, order.ErrEditConflict),
		errors.Is(err, order.ErrSurchargeRequired):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusConflict)
	case errors.Is(err, order.ErrNoLines):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnprocessableEntity)
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
	o.Post("/:number/refund", h.refundOrder)
	o.Get("/:number/refunds", h.listOrderRefunds)
	o.Post("/:number/items/preview", h.previewOrderItems)
	o.Put("/:number/items", h.updateOrderItems)
	o.Get("/:number/edits", h.listOrderEdits)
}
