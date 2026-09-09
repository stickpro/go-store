package order

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/models"
)

// linePricing carries the three price tiers of a product. resolveUnitPrice picks
// one based on the buyer.
type linePricing struct {
	retail    decimal.Decimal
	business  decimal.Decimal
	wholesale decimal.Decimal
}

// resolveUnitPrice returns the price a given buyer pays for a product.
//
// v1: always retail. There is no price group on the user yet; when it lands this
// switches on it. Guests (u == nil) also pay retail. Kept as the single choke
// point so cart and order agree.
func resolveUnitPrice(u *models.User, p linePricing) decimal.Decimal {
	_ = u
	return p.retail
}

// checkoutLine is one priced order line, assembled inside the checkout
// transaction before anything is written.
type checkoutLine struct {
	productID uuid.UUID
	variantID uuid.UUID
	sku       *string
	name      string
	slug      *string
	imagePath *string
	unitPrice decimal.Decimal
	quantity  int64

	// product parcel data (kg / cm), for shipping-cost calculation
	weightKG decimal.Decimal
	lengthCM decimal.Decimal
	widthCM  decimal.Decimal
	heightCM decimal.Decimal
}

func (l checkoutLine) lineTotal() decimal.Decimal {
	return l.unitPrice.Mul(decimal.NewFromInt(l.quantity))
}

type orderTotals struct {
	subtotal decimal.Decimal
	discount decimal.Decimal
	shipping decimal.Decimal
	tax      decimal.Decimal
	grand    decimal.Decimal
}

// computeTotals sums the lines and applies shipping. Discount and tax are zero
// in v1 (no promo / tax engine yet) but kept in the shape so the DB columns and
// the API contract don't change when they arrive.
func computeTotals(lines []checkoutLine, shipping decimal.Decimal) orderTotals {
	subtotal := decimal.Zero
	for _, l := range lines {
		subtotal = subtotal.Add(l.lineTotal())
	}

	t := orderTotals{
		subtotal: subtotal,
		discount: decimal.Zero,
		shipping: shipping,
		tax:      decimal.Zero,
	}
	t.grand = t.subtotal.Sub(t.discount).Add(t.shipping).Add(t.tax)
	return t
}

// shippingFor returns the shipping fee for a given subtotal: the configured flat
// amount, waived once the subtotal reaches the free-shipping threshold.
func (s *Service) shippingFor(subtotal decimal.Decimal) decimal.Decimal {
	if s.flatShipping.IsZero() {
		return decimal.Zero
	}
	if s.freeShippingFrom.IsPositive() && subtotal.GreaterThanOrEqual(s.freeShippingFrom) {
		return decimal.Zero
	}
	return s.flatShipping
}
