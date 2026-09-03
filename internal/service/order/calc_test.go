package order

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func line(price string, qty int64) checkoutLine {
	return checkoutLine{
		productID: uuid.New(),
		variantID: uuid.New(),
		unitPrice: dec(price),
		quantity:  qty,
	}
}

func TestComputeTotals(t *testing.T) {
	lines := []checkoutLine{
		line("10.50", 2), // 21.00
		line("4.25", 3),  // 12.75
	}

	got := computeTotals(lines, dec("5.00"))

	if !got.subtotal.Equal(dec("33.75")) {
		t.Fatalf("subtotal = %s, want 33.75", got.subtotal)
	}
	if !got.shipping.Equal(dec("5.00")) {
		t.Fatalf("shipping = %s, want 5.00", got.shipping)
	}
	if !got.grand.Equal(dec("38.75")) {
		t.Fatalf("grand = %s, want 38.75", got.grand)
	}
}

func TestComputeTotals_Empty(t *testing.T) {
	got := computeTotals(nil, decimal.Zero)
	if !got.grand.IsZero() || !got.subtotal.IsZero() {
		t.Fatalf("empty totals not zero: %+v", got)
	}
}

func TestShippingFor(t *testing.T) {
	cases := []struct {
		name     string
		flat     string
		freeFrom string
		subtotal string
		want     string
	}{
		{"no flat fee", "0", "0", "100", "0"},
		{"flat fee applies", "7.00", "0", "50", "7.00"},
		{"below free threshold", "7.00", "100", "99.99", "7.00"},
		{"at free threshold", "7.00", "100", "100", "0"},
		{"above free threshold", "7.00", "100", "250", "0"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &Service{flatShipping: dec(c.flat), freeShippingFrom: dec(c.freeFrom)}
			got := s.shippingFor(dec(c.subtotal))
			if !got.Equal(dec(c.want)) {
				t.Fatalf("shippingFor(%s) = %s, want %s", c.subtotal, got, c.want)
			}
		})
	}
}

func TestResolveUnitPrice_RetailForNow(t *testing.T) {
	p := linePricing{retail: dec("9.99"), business: dec("8.00"), wholesale: dec("6.00")}
	if got := resolveUnitPrice(nil, p); !got.Equal(dec("9.99")) {
		t.Fatalf("guest price = %s, want retail 9.99", got)
	}
}
