package config

// PaymentConfig holds the checkout payment-method catalogue, the way
// ShippingConfig.Methods holds the delivery one.
type PaymentConfig struct {
	// Methods is the checkout payment-method catalogue served by
	// GET /v1/payments/methods. Empty falls back to DefaultPaymentMethods.
	Methods []PaymentMethodConfig `yaml:"methods"`
}

// PaymentMethodConfig is one checkout payment method.
type PaymentMethodConfig struct {
	// Code is the stable id the frontend and checkout use (payment_method on
	// CreateOrderRequest, and the `provider` sent to POST /v1/orders/{id}/payment
	// for "online" methods).
	Code string `yaml:"code"`
	// Title is the human label shown on the method tab.
	Title string `yaml:"title"`
	// Kind: "cash" (settled offline — on delivery/pickup, nothing routed through
	// an acquirer) or "online" (routed through Provider via POST
	// /v1/orders/{id}/payment).
	Kind string `yaml:"kind"`
	// Provider is the acquiring provider code ("tbank", …) for an "online"
	// method; empty for "cash".
	Provider string `yaml:"provider"`
}

// DefaultPaymentMethods is used when payment.methods is empty.
func DefaultPaymentMethods() []PaymentMethodConfig {
	return []PaymentMethodConfig{
		{Code: "cash_on_delivery", Title: "Наличными при получении", Kind: "cash"},
		{Code: "tbank", Title: "Банковской картой онлайн", Kind: "online", Provider: "tbank"},
	}
}

// ResolvedMethods returns the configured methods, or the built-in defaults
// when none are set.
func (c PaymentConfig) ResolvedMethods() []PaymentMethodConfig {
	if len(c.Methods) == 0 {
		return DefaultPaymentMethods()
	}
	return c.Methods
}
