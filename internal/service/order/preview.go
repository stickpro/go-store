package order

import (
	"context"

	"github.com/stickpro/go-store/internal/dto"
)

// PreviewCheckout returns the server-computed money breakdown for the caller's
// cart under a delivery choice, without creating an order. The frontend renders
// GrandTotal directly instead of adding shipping to the cart total itself.
//
// Pricing here mirrors CreateOrder (same line prices, same resolveShipping) but
// runs outside the checkout transaction — it does not lock stock and is only
// indicative until the order is placed.
func (s *Service) PreviewCheckout(ctx context.Context, d dto.CheckoutPreviewDTO) (*dto.CheckoutPreviewResultDTO, error) {
	cart, err := s.cart.GetCart(ctx, d.Owner)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, ErrCartEmpty
	}

	lines := checkoutLinesFromCart(cart)

	sc, err := s.resolveShipping(ctx, d.Shipping, lines)
	if err != nil {
		return nil, err
	}
	totals := computeTotals(lines, sc.cost)

	itemCount := 0
	for _, l := range lines {
		itemCount += int(l.quantity)
	}

	return &dto.CheckoutPreviewResultDTO{
		Currency:      s.currency(),
		ItemCount:     itemCount,
		Subtotal:      totals.subtotal,
		DiscountTotal: totals.discount,
		ShippingTotal: totals.shipping,
		TaxTotal:      totals.tax,
		GrandTotal:    totals.grand,
		Shipping: dto.OrderShippingDTO{
			Postcode:   d.Shipping.Postcode,
			Method:     sc.method,
			Provider:   sc.provider,
			TariffCode: sc.tariff,
			PointCode:  sc.point,
			MinDays:    sc.minDays,
			MaxDays:    sc.maxDays,
		},
	}, nil
}

// checkoutLinesFromCart maps an enriched cart into priced checkout lines. Unlike
// buildOrder it does not lock rows or re-read the DB — the enriched cart already
// carries the retail price and the product's parcel dimensions.
func checkoutLinesFromCart(cart *dto.CartDTO) []checkoutLine {
	lines := make([]checkoutLine, 0, len(cart.Items))
	for _, it := range cart.Items {
		if !it.Available {
			continue
		}
		lines = append(lines, checkoutLine{
			productID: it.ProductID,
			variantID: it.VariantID,
			name:      it.Name,
			unitPrice: it.Price,
			quantity:  it.Quantity,
			weightKG:  it.WeightKG,
			lengthCM:  it.LengthCM,
			widthCM:   it.WidthCM,
			heightCM:  it.HeightCM,
		})
	}
	return lines
}
