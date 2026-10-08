package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/payment_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/payment_response"
	"github.com/stickpro/go-store/internal/delivery/middleware"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/order"
	"github.com/stickpro/go-store/internal/service/payment"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// initOrderPayment starts a payment attempt for an order through one provider.
//
//	@Summary		Start payment
//	@Description	Starts a payment attempt for the order at the chosen provider and returns a PaymentURL to redirect the customer to. `id` is the order's `id` field from the checkout response — not its human-facing `number`: unlike the sequential number, it isn't guessable, which is what lets a guest checkout (no account) start payment for its own order here.
//	@Tags			Payment
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Order id"
//	@Param			request	body		payment_request.InitPaymentRequest	true	"Provider choice"
//	@Success		200		{object}	response.Result[payment_response.PaymentResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/orders/{id}/payment [post]
func (h *Handler) initOrderPayment(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}

	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(errors.New("order id must be a uuid")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &payment_request.InitPaymentRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	o, err := h.services.OrderService.GetForPayment(c.Context(), owner, orderID)
	if err != nil {
		return h.orderError(err)
	}
	if o.PaymentStatus != constant.PaymentUnpaid.String() && o.PaymentStatus != constant.PaymentFailed.String() {
		return apierror.New().AddError(errors.New("order is already paid")).SetHttpCode(fiber.StatusConflict)
	}

	p, err := h.services.PaymentService.Init(c.Context(), dto.InitPaymentDTO{
		OrderID:       o.ID,
		OrderNumber:   o.Number,
		Provider:      req.Provider,
		Amount:        o.GrandTotal,
		Currency:      o.Currency,
		Description:   fmt.Sprintf("Оплата заказа №%d", o.Number),
		CustomerEmail: o.Email,
	})
	if err != nil {
		return h.paymentError(err)
	}

	return c.JSON(response.OkByData(payment_response.NewFromDTO(p)))
}

// getOrderPayment returns the latest payment attempt for an order.
//
//	@Summary		Get payment status
//	@Description	The most recent payment attempt for the order (a customer may retry after a failed/expired attempt). `id` is the order's `id`, not its `number` — see POST on the same path.
//	@Tags			Payment
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Order id"
//	@Success		200	{object}	response.Result[payment_response.PaymentResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Router			/v1/orders/{id}/payment [get]
func (h *Handler) getOrderPayment(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}

	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(errors.New("order id must be a uuid")).SetHttpCode(fiber.StatusBadRequest)
	}

	if _, err := h.services.OrderService.GetForPayment(c.Context(), owner, orderID); err != nil {
		return h.orderError(err)
	}

	p, err := h.services.PaymentService.GetLatestByOrderID(c.Context(), orderID)
	if err != nil {
		return h.paymentError(err)
	}

	return c.JSON(response.OkByData(payment_response.NewFromDTO(p)))
}

// paymentNotification receives a payment status webhook from a provider.
//
//	@Summary		Payment webhook
//	@Description	Provider webhook (T-Bank Notification API). Verifies the payload's signature and, on a confirmed payment, moves the order to "paid" (or "refunded" on a reversal). Not for browser/API clients — called by the provider only.
//	@Tags			Payment
//	@Accept			json
//	@Produce		plain
//	@Param			provider	path	string	true	"Provider code, e.g. tbank"
//	@Success		200
//	@Router			/v1/payments/{provider}/notification [post]
func (h *Handler) paymentNotification(c fiber.Ctx) error {
	providerCode := c.Params("provider")
	body := c.Body()

	// Logged unconditionally, before any parsing/validation: if webhooks
	// aren't showing up, this line is what tells us whether the request ever
	// reached the app at all (vs. a firewall/notification_url/DNS problem
	// upstream of it). Card data here is already masked by the provider
	// (T-Bank sends Pan like "430000******0000"), so logging the raw body is
	// safe — drop it to Debug once delivery is confirmed working.
	h.logger.Infow("payment: notification received",
		"provider", providerCode,
		"remote_ip", c.IP(),
		"body_size", len(body),
		"body", string(body),
	)

	p, ack, err := h.services.PaymentService.HandleNotification(c.Context(), providerCode, body)
	if err != nil {
		switch {
		case errors.Is(err, payment.ErrInvalidSignature):
			h.logger.Warnw("payment: invalid notification signature", "provider", providerCode, "remote_ip", c.IP())
			return c.Status(fiber.StatusBadRequest).SendString("invalid signature")
		case errors.Is(err, payment.ErrProviderNotFound):
			h.logger.Warnw("payment: notification for unknown provider", "provider", providerCode, "remote_ip", c.IP())
			return c.Status(fiber.StatusNotFound).SendString("unknown provider")
		case errors.Is(err, payment.ErrNotFound):
			// Nothing a retry would fix — ack so the provider stops resending it.
			h.logger.Warnw("payment: notification for unknown payment", "provider", providerCode)
			return c.SendString(ack)
		case errors.Is(err, payment.ErrAmountMismatch):
			// The payment stays unconfirmed and the order unpaid; an admin has
			// to look at it. Acked, since a resend would carry the same amount.
			h.logger.Errorw("payment: confirmed amount mismatch, payment left unconfirmed",
				"provider", providerCode, "payment_id", p.ID, "order_id", p.OrderID, "error", err)
			return c.SendString(ack)
		default:
			h.logger.Errorw("payment: handle notification", "provider", providerCode, "error", err)
			return c.Status(fiber.StatusInternalServerError).SendString("error")
		}
	}

	h.logger.Infow("payment: notification processed",
		"provider", providerCode,
		"payment_id", p.ID,
		"order_id", p.OrderID,
		"status", p.Status,
	)

	h.syncOrderWithPayment(c.Context(), p, providerCode)

	return c.SendString(ack)
}

// syncOrderWithPayment reconciles an order with a payment that just changed.
// Called after the webhook has already been acknowledged, so errors here are
// logged, not surfaced to the provider — an admin can always drive the same
// transition by hand if this fails.
func (h *Handler) syncOrderWithPayment(ctx context.Context, p *dto.PaymentDTO, providerCode string) {
	actor := constant.OrderActorPayment + ":" + providerCode

	switch payment.Status(p.Status) {
	case payment.StatusConfirmed:
		h.markOrderPaid(ctx, p, providerCode, actor)
	case payment.StatusRefunded, payment.StatusPartiallyRefunded:
		full := payment.Status(p.Status) == payment.StatusRefunded
		o, err := h.services.OrderService.RecordRefund(ctx, p.OrderID, p.RefundedAmount, full, actor)
		if err != nil {
			h.logger.Warnw("payment: record refund on order", "order_id", p.OrderID, "status", p.Status, "error", err)
			return
		}
		h.logger.Infow("payment: order refund recorded",
			"order", o.Number, "status", o.Status, "payment_status", o.PaymentStatus, "refunded_total", o.RefundedTotal)
	default:
		// Pending/intermediate status (e.g. T-Bank's AUTHORIZED before it
		// auto-confirms) — nothing for the order to do yet.
		h.logger.Infow("payment: notification status needs no order transition",
			"provider", providerCode, "payment_id", p.ID, "order_id", p.OrderID, "status", p.Status)
	}
}

func (h *Handler) markOrderPaid(ctx context.Context, p *dto.PaymentDTO, providerCode, actor string) {
	o, err := h.services.OrderService.GetByID(ctx, p.OrderID)
	if err != nil {
		h.logger.Errorw("payment: load order after notification", "order_id", p.OrderID, "error", err)
		return
	}
	if o.Status == constant.OrderPaid.String() {
		h.logger.Infow("payment: order already in target status, skipping",
			"order", o.Number, "status", o.Status)
		return
	}
	// The order may have been edited after this payment was started (the old
	// PaymentURL is voided on edit, but the customer can beat that). Money
	// taken for a different total must not mark the order paid; an admin
	// sorts it out — typically a refund and a new payment.
	if !p.Amount.Equal(o.GrandTotal) {
		h.logger.Errorw("payment: confirmed amount differs from order total, order left unpaid",
			"order", o.Number, "payment_id", p.ID, "paid", p.Amount, "grand_total", o.GrandTotal)
		return
	}

	method := providerCode
	if _, err := h.services.OrderService.UpdateStatus(ctx, o.Number, dto.OrderStatusUpdateDTO{
		Status:        constant.OrderPaid.String(),
		Actor:         actor,
		PaymentMethod: &method,
	}); err != nil {
		if errors.Is(err, order.ErrInvalidTransition) {
			h.logger.Warnw("payment: order not in a state to sync", "order", o.Number, "from", o.Status, "to", constant.OrderPaid, "error", err)
			return
		}
		h.logger.Errorw("payment: sync order status", "order", o.Number, "error", err)
		return
	}

	h.logger.Infow("payment: order status synced", "order", o.Number, "from", o.Status, "to", constant.OrderPaid)
}

// paymentError maps payment-service errors to HTTP responses.
func (h *Handler) paymentError(err error) error {
	switch {
	case errors.Is(err, payment.ErrProviderNotFound):
		return apierror.New().AddError(errors.New("unknown payment provider")).SetHttpCode(fiber.StatusUnprocessableEntity)
	case errors.Is(err, payment.ErrNotFound):
		return apierror.New().AddError(errors.New("payment not found")).SetHttpCode(fiber.StatusNotFound)
	case errors.Is(err, payment.ErrNotConfirmed):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusConflict)
	default:
		return h.handleError(err, "payment")
	}
}

// listPaymentMethods returns the checkout payment-method catalogue: one entry
// per method tab, "cash" (settled offline) or "online" (routed through an
// acquirer). Pass the entry's `code` back as payment_method at checkout; for
// an "online" entry, also pass its `provider` to POST /v1/orders/{id}/payment
// once the order exists.
//
//	@Summary	List payment methods
//	@Tags		Payment
//	@Produce	json
//	@Success	200	{object}	response.Result[[]payment_response.PaymentMethodResponse]
//	@Router		/v1/payments/methods [get]
func (h *Handler) listPaymentMethods(c fiber.Ctx) error {
	return c.JSON(response.OkByData(payment_response.NewMethodList(h.services.Payments.Methods())))
}

func (h *Handler) initPaymentRoutes(v1 fiber.Router) {
	o := v1.Group("/orders")
	o.Post("/:id/payment", middleware.OptionalAuthMiddleware(h.services.AuthService), h.initOrderPayment)
	o.Get("/:id/payment", middleware.OptionalAuthMiddleware(h.services.AuthService), h.getOrderPayment)

	p := v1.Group("/payments")
	p.Get("/methods", h.listPaymentMethods)
	p.Post("/:provider/notification", h.paymentNotification)
}
