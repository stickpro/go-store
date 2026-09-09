package pochta

import (
	"context"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/shipping"
)

// Quote calculates Russian Post shipping cost via the public tariff calculator.
func (p *provider) Quote(ctx context.Context, q shipping.RateQuery) ([]dto.ShippingRate, error) {
	if !p.Enabled() {
		return nil, shipping.ErrRatesNotSupported
	}

	fromIndex := digits(q.FromPostalCode)
	if fromIndex == "" {
		fromIndex = digits(p.origin)
	}
	toIndex, err := p.destinationIndex(q)
	if err != nil {
		return nil, err
	}
	if fromIndex == "" || toIndex == "" {
		return nil, fmt.Errorf("pochta: origin and destination postal codes are required")
	}

	res, err := p.client.calculateTariff(ctx, pochtaTariffQuery{
		Object:       p.cfg.TariffObjectCode,
		From:         fromIndex,
		To:           toIndex,
		WeightGrams:  q.Parcel.WeightGrams,
		LengthCM:     q.Parcel.LengthCM,
		WidthCM:      q.Parcel.WidthCM,
		HeightCM:     q.Parcel.HeightCM,
		PackType:     p.cfg.TariffPackType,
		SumOCKopecks: q.Parcel.DeclaredValueKopecks,
	})
	if err != nil {
		return nil, err
	}

	return []dto.ShippingRate{{
		Provider:     "pochta",
		TariffCode:   strconv.Itoa(p.cfg.TariffObjectCode),
		TariffName:   firstNonEmpty(res.Name, "Почта России"),
		DeliveryType: "pickup",
		Cost:         kopecksToRub(res.TotalKopecks),
		Currency:     "RUB",
		MinDays:      res.MinDays,
		MaxDays:      res.MaxDays,
		Details: map[string]any{
			"deadline":      res.Deadline,
			"cost_no_vat":   kopecksToRub(res.NoVATKopecks),
			"weight_billed": res.WeightBilled,
		},
	}}, nil
}

func (p *provider) destinationIndex(q shipping.RateQuery) (string, error) {
	if q.ToPointCode != "" {
		pt, ok := p.FindPoint(q.ToPointCode)
		if !ok {
			return "", fmt.Errorf("pochta: unknown pickup point %q", q.ToPointCode)
		}
		return digits(firstNonEmpty(pt.PostalCode, pt.Code)), nil
	}
	return digits(q.ToPostalCode), nil
}

func kopecksToRub(k int64) decimal.Decimal {
	return decimal.NewFromInt(k).Div(decimal.NewFromInt(100))
}

// digits keeps only the digits of a postal code ("115551" stays, "115 551" → "115551").
func digits(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			out = append(out, s[i])
		}
	}
	return string(out)
}
