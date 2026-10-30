package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/invoice_response"
	"github.com/stickpro/go-store/internal/delivery/middleware"
	"github.com/stickpro/go-store/internal/service/invoice"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// getOrderInvoice returns the active invoice of an order paid by invoice.
//
//	@Summary		Get order invoice
//	@Description	The active invoice of an order with payment_method of kind `invoice` (legal entities). `id` is the order's `id`, not its `number` — as for the payment endpoints, it authorizes guest requests. Download the document from GET /v1/orders/{id}/invoice/pdf.
//	@Tags			Order
//	@Produce		json
//	@Param			id	path		string	true	"Order id"
//	@Success		200	{object}	response.Result[invoice_response.InvoiceResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Router			/v1/orders/{id}/invoice [get]
func (h *Handler) getOrderInvoice(c fiber.Ctx) error {
	orderID, err := h.authorizeOrderByID(c)
	if err != nil {
		return err
	}

	inv, err := h.services.InvoiceService.GetActive(c.Context(), nil, orderID)
	if err != nil {
		return h.invoiceError(err)
	}
	return c.JSON(response.OkByData(invoice_response.NewFromDTO(inv)))
}

// getOrderInvoicePDF renders the active invoice of an order as PDF.
//
//	@Summary		Download order invoice
//	@Description	The active invoice of an order as a PDF document. `id` is the order's `id`.
//	@Tags			Order
//	@Produce		application/pdf
//	@Param			id	path		string	true	"Order id"
//	@Success		200	{file}		file
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		503	{object}	apierror.Errors
//	@Router			/v1/orders/{id}/invoice/pdf [get]
func (h *Handler) getOrderInvoicePDF(c fiber.Ctx) error {
	orderID, err := h.authorizeOrderByID(c)
	if err != nil {
		return err
	}

	pdf, err := h.services.InvoiceService.RenderPDF(c.Context(), orderID)
	if err != nil {
		return h.invoiceError(err)
	}
	return response.SendPDF(c, pdf.FileName, pdf.Content)
}

// authorizeOrderByID parses :id and checks the caller may see that order (same
// rule as the payment endpoints: the id is the capability for guests).
func (h *Handler) authorizeOrderByID(c fiber.Ctx) (uuid.UUID, error) {
	owner, err := h.cartOwner(c)
	if err != nil {
		return uuid.Nil, err
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, apierror.New().AddError(errors.New("order id must be a uuid")).SetHttpCode(fiber.StatusBadRequest)
	}
	if _, err := h.services.OrderService.GetForPayment(c.Context(), owner, orderID); err != nil {
		return uuid.Nil, h.orderError(err)
	}
	return orderID, nil
}

// invoiceError maps invoice-service errors to HTTP responses.
func (h *Handler) invoiceError(err error) error {
	switch {
	case errors.Is(err, invoice.ErrNotFound):
		return apierror.New().AddError(errors.New("order has no active invoice")).SetHttpCode(fiber.StatusNotFound)
	case errors.Is(err, invoice.ErrRenderUnavailable):
		h.logger.Errorw("invoice: render pdf", "error", err)
		return apierror.New().AddError(errors.New("invoice document is temporarily unavailable, retry later")).
			SetHttpCode(fiber.StatusServiceUnavailable)
	default:
		return h.handleError(err, "invoice")
	}
}

func (h *Handler) initInvoiceRoutes(v1 fiber.Router) {
	o := v1.Group("/orders")
	o.Get("/:id/invoice", middleware.OptionalAuthMiddleware(h.services.AuthService), h.getOrderInvoice)
	o.Get("/:id/invoice/pdf", middleware.OptionalAuthMiddleware(h.services.AuthService), h.getOrderInvoicePDF)
}
