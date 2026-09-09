package dto

import "github.com/shopspring/decimal"

// ShippingRate is one delivery option for a parcel: a carrier + tariff with its
// price and delivery window. Carrier-specific extras live in Details.
type ShippingRate struct {
	Provider     string          `json:"provider"`    // "cdek" | "yandex_delivery" | "pochta"
	TariffCode   string          `json:"tariff_code"` // carrier's tariff/service id
	TariffName   string          `json:"tariff_name"`
	DeliveryType string          `json:"delivery_type"` // "pickup" | "courier"
	Cost         decimal.Decimal `json:"cost"`          // total to charge the customer
	Currency     string          `json:"currency"`
	MinDays      int             `json:"min_days"`
	MaxDays      int             `json:"max_days"`
	Details      map[string]any  `json:"details,omitempty"`
}
