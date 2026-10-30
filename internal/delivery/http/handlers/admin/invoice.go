package admin

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/invoice_response"
	"github.com/stickpro/go-store/internal/service/invoice"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// getOrderInvoice returns the active invoice of an order paid by invoice.
//
//	@Summary		Get order invoice
//	@Description	The active invoice of an order with an `invoice` payment method. Reissued automatically when the order's items, totals or buyer change while unpaid; voided on cancel.
//	@Tags			Admin Order
//	@Produce		json
//	@Param			number	path		int	true	"Order number"
//	@Success		200		{object}	response.Result[invoice_response.InvoiceResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/admin/orders/{number}/invoice [get]
//	@Security		BearerAuth
func (h *Handler) getOrderInvoice(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}
	o, err := h.services.OrderService.GetByNumberAdmin(c.Context(), number)
	if err != nil {
		return h.orderError(err)
	}

	inv, err := h.services.InvoiceService.GetActive(c.Context(), nil, o.ID)
	if err != nil {
		return h.invoiceError(err)
	}
	return c.JSON(response.OkByData(invoice_response.NewFromDTO(inv)))
}

// getOrderInvoicePDF renders the active invoice of an order as PDF.
//
//	@Summary	Download order invoice
//	@Tags		Admin Order
//	@Produce	application/pdf
//	@Param		number	path		int	true	"Order number"
//	@Success	200		{file}		file
//	@Failure	400		{object}	apierror.Errors
//	@Failure	404		{object}	apierror.Errors
//	@Failure	503		{object}	apierror.Errors
//	@Router		/v1/admin/orders/{number}/invoice/pdf [get]
//	@Security	BearerAuth
func (h *Handler) getOrderInvoicePDF(c fiber.Ctx) error {
	number, err := parseOrderNumber(c)
	if err != nil {
		return err
	}
	o, err := h.services.OrderService.GetByNumberAdmin(c.Context(), number)
	if err != nil {
		return h.orderError(err)
	}

	pdf, err := h.services.InvoiceService.RenderPDF(c.Context(), o.ID)
	if err != nil {
		return h.invoiceError(err)
	}
	return response.SendPDF(c, pdf.FileName, pdf.Content)
}

func (h *Handler) invoiceError(err error) error {
	switch {
	case errors.Is(err, invoice.ErrNotFound):
		return apierror.New().AddError(errors.New("order has no active invoice")).SetHttpCode(fiber.StatusNotFound)
	case errors.Is(err, invoice.ErrRenderUnavailable):
		h.logger.Errorw("admin: render invoice pdf", "error", err)
		return apierror.New().AddError(errors.New("invoice renderer is unavailable")).SetHttpCode(fiber.StatusServiceUnavailable)
	default:
		return h.handleError(err, "invoice")
	}
}
