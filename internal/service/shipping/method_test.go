package shipping

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
)

func dm(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestMethodFinalCost(t *testing.T) {
	tests := []struct {
		markup string
		base   string
		want   string
	}{
		{"", "267.50", "268"},   // no markup -> just round up
		{"0", "300", "300"},     // exact
		{"10", "267.50", "295"}, // 267.50 * 1.10 = 294.25 -> 295
		{"15", "420.00", "483"}, // 420 * 1.15 = 483
		{"7.5", "1000", "1075"}, // 1000 * 1.075 = 1075
	}
	for _, tt := range tests {
		m := Method{MarkupPercent: dmOrZero(tt.markup)}
		if got := m.FinalCost(dm(tt.base)); !got.Equal(dm(tt.want)) {
			t.Fatalf("markup %q base %s: got %s, want %s", tt.markup, tt.base, got, tt.want)
		}
	}
}

func dmOrZero(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

func TestMethodsFromConfigParsesMarkup(t *testing.T) {
	ms := methodsFromConfig([]config.ShippingMethodConfig{
		{Code: "cdek", Provider: "cdek", Tariff: "136", Markup: "12.5"},
		{Code: "pickup", Kind: "self_pickup", Free: true},
	})
	if !ms[0].MarkupPercent.Equal(dm("12.5")) {
		t.Fatalf("markup not parsed: %s", ms[0].MarkupPercent)
	}
	if !ms[1].MarkupPercent.IsZero() {
		t.Fatalf("empty markup should be zero: %s", ms[1].MarkupPercent)
	}
}
