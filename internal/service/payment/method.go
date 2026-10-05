package payment

import "github.com/stickpro/go-store/internal/config"

// MethodKind classifies a checkout payment method.
type MethodKind string //	@name	PaymentMethodKind

const (
	// MethodCash is settled offline — cash (or a card terminal) on delivery or
	// pickup. Nothing is routed through an acquirer; the order is simply
	// stamped with this code as its payment_method at checkout.
	MethodCash MethodKind = "cash"
	// MethodOnline is routed through Provider: after checkout, the frontend
	// calls POST /v1/orders/{id}/payment with this method's Provider code.
	MethodOnline MethodKind = "online"
)

// Method is one checkout payment option: a stable code the frontend and
// checkout use. Mirrors shipping.Method.
type Method struct {
	Code     string
	Title    string
	Kind     MethodKind
	Provider string
}

// MethodInfo is a Method plus the availability flag the frontend needs to
// render its method tabs.
type MethodInfo struct {
	Method
	// Enabled: false hides the tab. Always true for MethodCash; for
	// MethodOnline, true only if Provider is registered and enabled.
	Enabled bool
}

func methodsFromConfig(cfgs []config.PaymentMethodConfig) []Method {
	out := make([]Method, 0, len(cfgs))
	for _, c := range cfgs {
		kind := MethodKind(c.Kind)
		if kind == "" {
			kind = MethodCash
		}
		out = append(out, Method{
			Code:     c.Code,
			Title:    c.Title,
			Kind:     kind,
			Provider: c.Provider,
		})
	}
	return out
}
