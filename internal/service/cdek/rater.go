package cdek

import (
	"context"
	"fmt"
	"strconv"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/shipping"
)

// cdekTariffNames maps the common CDEK tariff codes to a human label for the UI.
var cdekTariffNames = map[int]string{
	136: "Посылка склад-склад",
	137: "Посылка склад-дверь",
	138: "Посылка дверь-склад",
	139: "Посылка дверь-дверь",
	366: "Экспресс склад-склад",
	368: "Экспресс склад-дверь",
	234: "Экономичная посылка склад-склад",
	233: "Экономичная посылка склад-дверь",
}

// Quote calculates CDEK shipping cost for every configured tariff code.
func (p *provider) Quote(ctx context.Context, q shipping.RateQuery) ([]dto.ShippingRate, error) {
	if !p.Enabled() || len(p.cfg.TariffCodes) == 0 {
		return nil, shipping.ErrRatesNotSupported
	}

	fromIndex := q.FromPostalCode
	if fromIndex == "" {
		fromIndex = p.origin
	}
	toIndex, err := p.destinationIndex(q)
	if err != nil {
		return nil, err
	}
	if fromIndex == "" || toIndex == "" {
		return nil, fmt.Errorf("cdek: origin and destination postal codes are required")
	}

	rates := make([]dto.ShippingRate, 0, len(p.cfg.TariffCodes))
	for _, code := range p.cfg.TariffCodes {
		res, err := p.client.calculateTariff(ctx, cdekTariffRequest{
			Type:         1, // интернет-магазин
			TariffCode:   code,
			FromLocation: cdekTariffLocation{PostalCode: fromIndex},
			ToLocation:   cdekTariffLocation{PostalCode: toIndex},
			Packages: []cdekTariffPackage{{
				Weight: q.Parcel.WeightGrams,
				Length: q.Parcel.LengthCM,
				Width:  q.Parcel.WidthCM,
				Height: q.Parcel.HeightCM,
			}},
		})
		if err != nil {
			p.client.logger.Errorw("cdek: tariff calculation failed", "tariff_code", code, "error", err)
			continue
		}

		rates = append(rates, dto.ShippingRate{
			Provider:     "cdek",
			TariffCode:   strconv.Itoa(code),
			TariffName:   firstNonEmpty(cdekTariffNames[code], "Тариф "+strconv.Itoa(code)),
			DeliveryType: tariffDeliveryType(code),
			Cost:         res.TotalSum,
			Currency:     firstNonEmpty(res.Currency, "RUB"),
			MinDays:      res.PeriodMin,
			MaxDays:      res.PeriodMax,
			Details: map[string]any{
				"weight_calc":   res.WeightCalc,
				"delivery_mode": res.DeliveryMode,
			},
		})
	}
	if len(rates) == 0 {
		return nil, fmt.Errorf("cdek: no tariff could be calculated")
	}
	return rates, nil
}

// destinationIndex resolves the delivery destination to a postal index: either
// the request's ToPostalCode or the postal code of the chosen pickup point.
func (p *provider) destinationIndex(q shipping.RateQuery) (string, error) {
	if q.ToPointCode != "" {
		pt, ok := p.FindPoint(q.ToPointCode)
		if !ok {
			return "", fmt.Errorf("cdek: unknown pickup point %q", q.ToPointCode)
		}
		return pt.PostalCode, nil
	}
	return q.ToPostalCode, nil
}

// tariffDeliveryType classifies a CDEK tariff code as pickup or courier by its
// "-склад" (warehouse) vs "-дверь" (door) destination half.
func tariffDeliveryType(code int) string {
	switch code {
	case 137, 139, 368, 233:
		return "courier"
	default:
		return "pickup"
	}
}
