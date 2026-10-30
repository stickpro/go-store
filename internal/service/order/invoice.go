package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/invoice"
	"github.com/stickpro/go-store/internal/service/mail"
	"github.com/stickpro/go-store/internal/service/payment"
)

// checkPaymentMethod enforces which payment methods fit the buyer: a legal
// entity pays by invoice or online, and only with an available method; the
// invoice method is for legal entities only. Unknown codes from a private
// person are let through, as before.
func (s *Service) checkPaymentMethod(customerType, code string) error {
	if s.payments == nil {
		return nil
	}
	m, ok := s.payments.Available(code)
	if customerType == constant.CustomerLegal {
		if !ok || (m.Kind != payment.MethodInvoice && m.Kind != payment.MethodOnline) {
			return fmt.Errorf("%w: %q for a legal entity (use an invoice or online method)", ErrPaymentMethodNotAllowed, code)
		}
		return nil
	}
	if m, known := s.payments.Method(code); known && m.Kind == payment.MethodInvoice {
		return fmt.Errorf("%w: %q requires customer_type legal", ErrPaymentMethodNotAllowed, code)
	}
	return nil
}

// wantsInvoice reports whether the order should carry an active invoice: a
// legal buyer, the invoice payment method, and still awaiting payment.
func (s *Service) wantsInvoice(o *dto.OrderDTO) bool {
	if s.invoices == nil || s.payments == nil || o.LegalEntity == nil || o.PaymentMethod == nil {
		return false
	}
	switch constant.OrderStatus(o.Status) {
	case constant.OrderNew, constant.OrderPending:
	default:
		return false
	}
	m, ok := s.payments.Available(*o.PaymentMethod)
	return ok && m.Kind == payment.MethodInvoice
}

// awaitingPayment reports whether no money has been taken on the order yet.
func awaitingPayment(o *dto.OrderDTO) bool {
	return o.PaymentStatus == constant.PaymentUnpaid.String() || o.PaymentStatus == constant.PaymentFailed.String()
}

// syncInvoice brings the order's invoice in line with the order inside the
// transaction that changed it: issues or reissues it while the order awaits
// payment by invoice, voids it when it no longer applies. Once paid, the
// invoice the customer paid stays as it is. Returns the active invoice, if any.
func (s *Service) syncInvoice(ctx context.Context, tx pgx.Tx, o *dto.OrderDTO) (*dto.InvoiceDTO, error) {
	if s.invoices == nil {
		return nil, nil //nolint:nilnil // no invoicing wired up
	}

	if !awaitingPayment(o) {
		inv, err := s.invoices.GetActive(ctx, tx, o.ID)
		if errors.Is(err, invoice.ErrNotFound) {
			return nil, nil //nolint:nilnil // paid without an invoice
		}
		return inv, err
	}

	if s.wantsInvoice(o) {
		inv, err := s.invoices.Ensure(ctx, tx, o)
		if err != nil {
			return nil, fmt.Errorf("order: issue invoice: %w", err)
		}
		return inv, nil
	}

	if err := s.invoices.Void(ctx, tx, o.ID); err != nil {
		return nil, fmt.Errorf("order: void invoice: %w", err)
	}
	return nil, nil //nolint:nilnil // no invoice applies
}

// location is the store timezone dates are shown in; UTC when unset or invalid.
func (s *Service) location() *time.Location {
	if s.cfg == nil || s.cfg.Order.Timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(s.cfg.Order.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

// addInvoiceToConfirmation puts the invoice number and due date into the
// order confirmation email of an order paid by invoice.
func (s *Service) addInvoiceToConfirmation(ctx context.Context, o *dto.OrderDTO, msg *mail.OrderConfirmation) {
	if !s.wantsInvoice(o) {
		return
	}
	inv, err := s.invoices.GetActive(ctx, nil, o.ID)
	if err != nil {
		s.logger.Errorw("order: load invoice for confirmation email", "order", o.Number, "error", err)
		return
	}
	msg.InvoiceNumber = inv.Number
	msg.InvoiceDueDate = inv.DueAt.In(s.location()).Format("02.01.2006")
}
