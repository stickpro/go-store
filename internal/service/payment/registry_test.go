package payment

import (
	"context"
	"testing"

	"github.com/stickpro/go-store/internal/config"
)

// stubProvider is a Provider that's just a code + enabled flag — enough to
// drive Registry.Get/Methods without a real gateway.
type stubProvider struct {
	code    string
	enabled bool
}

func (s stubProvider) Code() string  { return s.code }
func (s stubProvider) Enabled() bool { return s.enabled }
func (s stubProvider) Init(context.Context, InitRequest) (InitResult, error) {
	return InitResult{}, nil
}

func (s stubProvider) HandleNotification(context.Context, []byte) (NotificationResult, error) {
	return NotificationResult{}, nil
}

func (s stubProvider) Cancel(context.Context, CancelRequest) (CancelResult, error) {
	return CancelResult{}, nil
}

func TestRegistryGet(t *testing.T) {
	reg := NewRegistry(nil,
		stubProvider{code: "tbank", enabled: true},
		stubProvider{code: "yookassa", enabled: false},
	)

	if _, ok := reg.Get("tbank"); !ok {
		t.Fatal("tbank should be enabled")
	}
	if _, ok := reg.Get("yookassa"); ok {
		t.Fatal("yookassa is disabled, Get should report not found")
	}
	if _, ok := reg.Get("missing"); ok {
		t.Fatal("unregistered provider")
	}
}

func TestMethodsCatalogue(t *testing.T) {
	reg := NewRegistry(
		[]config.PaymentMethodConfig{
			{Code: "cash_on_delivery", Title: "Наличными", Kind: "cash"},
			{Code: "tbank", Title: "Картой онлайн", Kind: "online", Provider: "tbank"},
			{Code: "off", Title: "Отключён", Kind: "online", Provider: "missing"},
			{Code: "disabled_provider", Title: "Выключенный провайдер", Kind: "online", Provider: "yookassa"},
		},
		stubProvider{code: "tbank", enabled: true},
		stubProvider{code: "yookassa", enabled: false},
	)

	got := map[string]MethodInfo{}
	for _, m := range reg.Methods() {
		got[m.Code] = m
	}

	if m := got["cash_on_delivery"]; !m.Enabled || m.Kind != MethodCash {
		t.Fatalf("cash_on_delivery: %+v", m)
	}
	if m := got["tbank"]; !m.Enabled || m.Kind != MethodOnline {
		t.Fatalf("tbank: %+v", m)
	}
	if m := got["off"]; m.Enabled {
		t.Fatalf("unknown provider must be disabled: %+v", m)
	}
	if m := got["disabled_provider"]; m.Enabled {
		t.Fatalf("disabled provider must be disabled: %+v", m)
	}

	if m, ok := reg.Method("tbank"); !ok || m.Provider != "tbank" {
		t.Fatalf("Method(tbank): %+v ok=%v", m, ok)
	}
	if _, ok := reg.Method("missing"); ok {
		t.Fatal("unregistered method code")
	}
}
