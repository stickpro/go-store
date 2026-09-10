package shipping

import (
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
)

// MethodKind classifies a checkout delivery method.
type MethodKind string

const (
	MethodSelfPickup MethodKind = "self_pickup" // from the store
	MethodPickup     MethodKind = "pickup"      // a carrier pickup point
	MethodCourier    MethodKind = "courier"     // a carrier delivers to an address
)

// Method is one checkout delivery option: a stable code the frontend and
// checkout use, mapped to a carrier + tariff so the frontend never touches
// tariff codes.
type Method struct {
	Code       string
	Title      string
	Kind       MethodKind
	Provider   string
	TariffCode string
	Free       bool
	// MarkupPercent is added on top of the carrier quote for this method.
	MarkupPercent decimal.Decimal
}

// FinalCost applies this method's markup to a carrier's base cost and rounds the
// result up to a whole currency unit.
func (m Method) FinalCost(base decimal.Decimal) decimal.Decimal {
	if m.MarkupPercent.IsPositive() {
		factor := decimal.NewFromInt(1).Add(m.MarkupPercent.Div(decimal.NewFromInt(100)))
		base = base.Mul(factor)
	}
	return base.Ceil()
}

// MethodInfo is a Method plus the availability flags the frontend needs to
// render its method tabs.
type MethodInfo struct {
	Method
	// Enabled: false hides the tab (carrier disabled in config).
	Enabled bool
	// HasPoints: the frontend should show a pickup-point map
	// (GET /v1/delivery/{provider}/points).
	HasPoints bool
	// HasRates: the carrier can quote a price; when false the frontend shows no
	// cost and checkout falls back to flat shipping.
	HasRates bool
}

func methodsFromConfig(cfgs []config.ShippingMethodConfig) []Method {
	out := make([]Method, 0, len(cfgs))
	for _, c := range cfgs {
		kind := MethodKind(c.Kind)
		if kind == "" {
			kind = MethodPickup
		}
		markup, _ := decimal.NewFromString(c.Markup)
		out = append(out, Method{
			Code:          c.Code,
			Title:         c.Title,
			Kind:          kind,
			Provider:      c.Provider,
			TariffCode:    c.Tariff,
			Free:          c.Free,
			MarkupPercent: markup,
		})
	}
	return out
}
