// Package tbank is the T-Bank (Т-Банк) acquiring provider: a thin,
// stateless implementation of payment.Provider on top of the Init/
// Notification v2 API. It never touches the database — persistence and
// orchestration belong to payment.Service.
package tbank

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/payment"
)

type provider struct {
	cfg    config.TBankConfig
	client *client
}

// New builds the T-Bank acquiring provider.
func New(cfg *config.Config) payment.Provider {
	return &provider{
		cfg:    cfg.TBank,
		client: newClient(cfg.TBank),
	}
}

func (p *provider) Code() string  { return "tbank" }
func (p *provider) Enabled() bool { return p.cfg.Enabled }

func (p *provider) Init(ctx context.Context, r payment.InitRequest) (payment.InitResult, error) {
	resp, raw, err := p.client.init(ctx, initCall{
		// Our own payments.id, not T-Bank's PaymentId (unknown until this
		// call returns) — HandleNotification matches webhooks back by this.
		OrderID:         r.PaymentID.String(),
		AmountKopecks:   toKopecks(r.Amount),
		Description:     r.Description,
		NotificationURL: p.cfg.NotificationURL,
		SuccessURL:      withOrderParam(p.cfg.SuccessURL, r.OrderNumber),
		FailURL:         withOrderParam(p.cfg.FailURL, r.OrderNumber),
	})
	if err != nil {
		return payment.InitResult{}, err
	}

	return payment.InitResult{
		ProviderPaymentID: resp.PaymentId.String(),
		PaymentURL:        resp.PaymentURL,
		Status:            mapStatus(resp.Status),
		RawResponse:       raw,
	}, nil
}

func (p *provider) HandleNotification(_ context.Context, raw []byte) (payment.NotificationResult, error) {
	n, valid, err := verifyNotification(p.cfg.Password, raw)
	if err != nil {
		return payment.NotificationResult{}, fmt.Errorf("tbank: parse notification: %w", err)
	}
	if !valid {
		return payment.NotificationResult{}, payment.ErrInvalidSignature
	}

	paymentID, err := uuid.Parse(n.OrderId)
	if err != nil {
		return payment.NotificationResult{}, fmt.Errorf("tbank: notification OrderId %q is not a payment id: %w", n.OrderId, err)
	}

	return payment.NotificationResult{
		PaymentID:         paymentID,
		ProviderPaymentID: n.PaymentId.String(),
		Status:            mapStatus(n.Status),
		RawNotification:   raw,
		// T-Bank stops retrying a webhook only once it gets back exactly "OK".
		AckBody: "OK",
	}, nil
}

func (p *provider) Cancel(ctx context.Context, r payment.CancelRequest) (payment.CancelResult, error) {
	resp, raw, err := p.client.cancel(ctx, r.ProviderPaymentID, toKopecks(r.Amount))
	if err != nil {
		return payment.CancelResult{}, err
	}

	return payment.CancelResult{
		Status:      mapStatus(resp.Status),
		RawResponse: raw,
	}, nil
}

// withOrderParam appends ?order=<number> (or &order=<number> if rawURL
// already has a query string) to the configured success/fail URL, so the
// page the customer lands on after paying knows which order to show —
// without this, SuccessURL/FailURL are the same static page for every order.
// An empty or unparseable rawURL is returned unchanged (T-Bank treats an
// empty SuccessURL/FailURL as "use the terminal's default").
func withOrderParam(rawURL string, orderNumber int64) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set("order", strconv.FormatInt(orderNumber, 10))
	u.RawQuery = q.Encode()
	return u.String()
}

// toKopecks converts a decimal ruble amount to the integer kopecks T-Bank's
// API expects. A zero/absent amount (e.g. "full refund") stays 0.
func toKopecks(amount decimal.Decimal) int64 {
	return amount.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
}

// mapStatus normalizes a T-Bank payment status onto payment.Status. Anything
// not explicitly terminal (AUTHORIZING, FORM_SHOWED, AUTHORIZED, …) is
// reported as pending — one-stage payments (PayType=O) move straight from
// AUTHORIZED to CONFIRMED on T-Bank's side without another call from us.
func mapStatus(s string) payment.Status {
	switch s {
	case statusConfirmed:
		return payment.StatusConfirmed
	case statusRejected, statusDeadlineExpired, statusCanceled:
		return payment.StatusFailed
	case statusReversed, statusPartialReversed, statusRefunded, statusPartialRefunded:
		return payment.StatusRefunded
	default:
		return payment.StatusPending
	}
}

var _ payment.Provider = (*provider)(nil)
