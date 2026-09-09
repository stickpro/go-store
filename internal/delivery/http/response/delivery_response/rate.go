package delivery_response

import (
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

// ShippingRateResponse is one delivery option for a parcel.
type ShippingRateResponse struct {
	Provider     string          `json:"provider"`
	TariffCode   string          `json:"tariff_code"`
	TariffName   string          `json:"tariff_name"`
	DeliveryType string          `json:"delivery_type"`
	Cost         decimal.Decimal `json:"cost"`
	Currency     string          `json:"currency"`
	MinDays      int             `json:"min_days"`
	MaxDays      int             `json:"max_days"`
	Details      map[string]any  `json:"details,omitempty"`
} //	@name	ShippingRateResponse

func NewRateFromDTO(r dto.ShippingRate) *ShippingRateResponse {
	return &ShippingRateResponse{
		Provider:     r.Provider,
		TariffCode:   r.TariffCode,
		TariffName:   r.TariffName,
		DeliveryType: r.DeliveryType,
		Cost:         r.Cost,
		Currency:     r.Currency,
		MinDays:      r.MinDays,
		MaxDays:      r.MaxDays,
		Details:      r.Details,
	}
}

func NewRateList(rates []dto.ShippingRate) []*ShippingRateResponse {
	out := make([]*ShippingRateResponse, 0, len(rates))
	for _, r := range rates {
		out = append(out, NewRateFromDTO(r))
	}
	return out
}
