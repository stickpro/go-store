package order

import (
	"context"
	"errors"
	"testing"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/invoice"
	"github.com/stickpro/go-store/internal/service/payment"
)

type onlineProvider struct{}

func (onlineProvider) Code() string  { return "tbank" }
func (onlineProvider) Enabled() bool { return true }
func (onlineProvider) Init(context.Context, payment.InitRequest) (payment.InitResult, error) {
	return payment.InitResult{}, nil
}

func (onlineProvider) HandleNotification(context.Context, []byte) (payment.NotificationResult, error) {
	return payment.NotificationResult{}, nil
}

func (onlineProvider) Cancel(context.Context, payment.CancelRequest) (payment.CancelResult, error) {
	return payment.CancelResult{}, nil
}

func registryWithInvoices(enabled bool) *payment.Registry {
	return payment.NewRegistry(config.DefaultPaymentMethods(), onlineProvider{}).WithInvoices(enabled)
}

func TestCheckPaymentMethod(t *testing.T) {
	s := &Service{payments: registryWithInvoices(true)}
	cases := []struct {
		customer, method string
		selfPickup       bool
		ok               bool
	}{
		// Legal entity: invoice only, whatever the delivery method.
		{constant.CustomerLegal, "invoice", false, true},
		{constant.CustomerLegal, "invoice", true, true},
		{constant.CustomerLegal, "tbank", false, false},
		{constant.CustomerLegal, "cash_on_delivery", true, false},
		{constant.CustomerLegal, "unknown", false, false},
		// Private person: cash only on store self-pickup, card otherwise.
		{constant.CustomerIndividual, "invoice", false, false},
		{constant.CustomerIndividual, "tbank", false, true},
		{constant.CustomerIndividual, "tbank", true, false},
		{constant.CustomerIndividual, "cash_on_delivery", true, true},
		{constant.CustomerIndividual, "cash_on_delivery", false, false},
		{constant.CustomerIndividual, "unknown", false, true}, // unchanged legacy behaviour
	}
	for _, c := range cases {
		err := s.checkPaymentMethod(c.customer, c.method, c.selfPickup)
		if c.ok && err != nil {
			t.Errorf("%s/%s/selfPickup=%v: unexpected error %v", c.customer, c.method, c.selfPickup, err)
		}
		if !c.ok && !errors.Is(err, ErrPaymentMethodNotAllowed) {
			t.Errorf("%s/%s/selfPickup=%v: want ErrPaymentMethodNotAllowed, got %v", c.customer, c.method, c.selfPickup, err)
		}
	}

	// Seller requisites not configured: the invoice method is unavailable.
	s = &Service{payments: registryWithInvoices(false)}
	if err := s.checkPaymentMethod(constant.CustomerLegal, "invoice", false); !errors.Is(err, ErrPaymentMethodNotAllowed) {
		t.Errorf("invoice without seller config: want ErrPaymentMethodNotAllowed, got %v", err)
	}
}

// invoiceStub satisfies the interface; wantsInvoice never calls it.
type invoiceStub struct{ invoice.IInvoiceService }

func TestWantsInvoice(t *testing.T) {
	s := &Service{payments: registryWithInvoices(true), invoices: invoiceStub{}}
	method := "invoice"
	card := "tbank"
	legal := &dto.LegalEntityDTO{Name: "x", INN: "7707083893", Address: "y"}

	cases := []struct {
		name string
		o    dto.OrderDTO
		want bool
	}{
		{"legal pending invoice", dto.OrderDTO{Status: "pending", LegalEntity: legal, PaymentMethod: &method}, true},
		{"legal pending card", dto.OrderDTO{Status: "pending", LegalEntity: legal, PaymentMethod: &card}, false},
		{"individual invoice", dto.OrderDTO{Status: "pending", PaymentMethod: &method}, false},
		{"legal paid", dto.OrderDTO{Status: "paid", LegalEntity: legal, PaymentMethod: &method}, false},
		{"legal cancelled", dto.OrderDTO{Status: "cancelled", LegalEntity: legal, PaymentMethod: &method}, false},
	}
	for _, c := range cases {
		if got := s.wantsInvoice(&c.o); got != c.want {
			t.Errorf("%s: wantsInvoice = %v, want %v", c.name, got, c.want)
		}
	}
}
