package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/shipping"
)

// ErrShippingUnavailable means the carrier option the customer picked is no
// longer offered for this parcel/route (price changed, tariff withdrawn, …).
var ErrShippingUnavailable = errors.New("order: chosen shipping option is unavailable")

// ErrShippingMethodUnknown means delivery_method_code is not in the configured
// method catalogue.
var ErrShippingMethodUnknown = errors.New("order: unknown delivery method")

// shippingChoice is the checkout's resolved shipping: the cost that goes into
// the order total plus the carrier snapshot stored on the order.
type shippingChoice struct {
	cost     decimal.Decimal
	method   *string
	provider *string
	tariff   *string
	point    *string
	minDays  *int32
	maxDays  *int32
}

// resolveShipping prices delivery. When the customer picked a carrier + tariff
// it re-quotes that option live and uses the server-computed cost (never the
// client's); otherwise it falls back to the configured flat/free shipping.
func (s *Service) resolveShipping(ctx context.Context, sel dto.ShippingSelection, lines []checkoutLine) (shippingChoice, error) {
	flat := func() shippingChoice {
		return shippingChoice{cost: s.shippingFor(sumSubtotal(lines)), method: sel.Method}
	}

	// A delivery_method_code resolves to a carrier + tariff (or a free / store
	// pickup with no carrier at all).
	if sel.MethodCode != nil && *sel.MethodCode != "" {
		if s.shipping == nil {
			return flat(), nil
		}
		m, ok := s.shipping.Method(*sel.MethodCode)
		if !ok {
			return shippingChoice{}, ErrShippingMethodUnknown
		}
		if sel.Method == nil {
			sel.Method = strPtrOrNil(m.Title)
		}
		if m.Free || m.Kind == shipping.MethodSelfPickup || m.Provider == "" {
			return shippingChoice{cost: decimal.Zero, method: sel.Method}, nil
		}
		provider, tariff := m.Provider, m.TariffCode
		sel.Provider, sel.TariffCode = &provider, &tariff
	}

	if s.shipping == nil || sel.Provider == nil || *sel.Provider == "" || sel.TariffCode == nil {
		return flat(), nil
	}

	parcel := shipping.ParcelFromItems(toParcelItems(lines), s.shipping.ParcelDefaults())
	rates, err := s.shipping.Quote(ctx, *sel.Provider, shipping.RateQuery{
		ToPostalCode: deref(sel.Postcode),
		ToPointCode:  deref(sel.PointCode),
		Parcel:       parcel,
	})
	if err != nil {
		return shippingChoice{}, fmt.Errorf("order: quote %s: %w", *sel.Provider, err)
	}

	for _, r := range rates {
		if r.TariffCode != *sel.TariffCode {
			continue
		}
		r := r
		return shippingChoice{
			cost:     r.Cost,
			method:   strPtrOrNil(r.TariffName),
			provider: strPtrOrNil(r.Provider),
			tariff:   strPtrOrNil(r.TariffCode),
			point:    sel.PointCode,
			minDays:  int32PtrOrNil(r.MinDays),
			maxDays:  int32PtrOrNil(r.MaxDays),
		}, nil
	}
	return shippingChoice{}, ErrShippingUnavailable
}

func toParcelItems(lines []checkoutLine) []shipping.ParcelItem {
	items := make([]shipping.ParcelItem, len(lines))
	for i, l := range lines {
		items[i] = shipping.ParcelItem{
			Quantity: l.quantity,
			Price:    l.unitPrice,
			WeightKG: l.weightKG,
			LengthCM: l.lengthCM,
			WidthCM:  l.widthCM,
			HeightCM: l.heightCM,
		}
	}
	return items
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func firstNonNil(a, b *string) *string {
	if a != nil {
		return a
	}
	return b
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int32PtrOrNil(n int) *int32 {
	if n == 0 {
		return nil
	}
	v := int32(n)
	return &v
}
