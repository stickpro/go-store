package payment

import "github.com/stickpro/go-store/internal/config"

// Registry is the set of acquiring providers this store wires up, indexed by
// code, plus the checkout payment-method catalogue. It is the payment-side
// counterpart of shipping.Registry.
type Registry struct {
	byCode    map[string]Provider
	methods   []Method
	methodIdx map[string]Method
}

// NewRegistry indexes providers by their Code and builds the payment-method
// catalogue.
func NewRegistry(methods []config.PaymentMethodConfig, providers ...Provider) *Registry {
	byCode := make(map[string]Provider, len(providers))
	for _, p := range providers {
		byCode[p.Code()] = p
	}

	ms := methodsFromConfig(methods)
	methodIdx := make(map[string]Method, len(ms))
	for _, m := range ms {
		methodIdx[m.Code] = m
	}

	return &Registry{byCode: byCode, methods: ms, methodIdx: methodIdx}
}

// Get returns the provider registered under code, if it exists and is
// enabled in config.
func (r *Registry) Get(code string) (Provider, bool) {
	p, ok := r.byCode[code]
	if !ok || !p.Enabled() {
		return nil, false
	}
	return p, true
}

// All returns every registered provider, enabled or not.
func (r *Registry) All() []Provider {
	out := make([]Provider, 0, len(r.byCode))
	for _, p := range r.byCode {
		out = append(out, p)
	}
	return out
}

// Method returns the catalogue entry for code.
func (r *Registry) Method(code string) (Method, bool) {
	m, ok := r.methodIdx[code]
	return m, ok
}

// Methods returns the payment-method catalogue with per-method availability
// resolved.
func (r *Registry) Methods() []MethodInfo {
	out := make([]MethodInfo, 0, len(r.methods))
	for _, m := range r.methods {
		out = append(out, r.methodInfo(m))
	}
	return out
}

func (r *Registry) methodInfo(m Method) MethodInfo {
	if m.Kind == MethodCash || m.Provider == "" {
		return MethodInfo{Method: m, Enabled: true}
	}
	_, ok := r.Get(m.Provider)
	return MethodInfo{Method: m, Enabled: ok}
}
