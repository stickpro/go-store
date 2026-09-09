package order_response

import (
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

// CheckoutPreviewResponse is the authoritative money breakdown for a cart under
// a delivery choice. Render grand_total directly.
type CheckoutPreviewResponse struct {
	Currency      string                `json:"currency"`
	ItemCount     int                   `json:"item_count"`
	Subtotal      decimal.Decimal       `json:"subtotal"`
	DiscountTotal decimal.Decimal       `json:"discount_total"`
	ShippingTotal decimal.Decimal       `json:"shipping_total"`
	TaxTotal      decimal.Decimal       `json:"tax_total"`
	GrandTotal    decimal.Decimal       `json:"grand_total"`
	Shipping      OrderShippingResponse `json:"shipping"`
} //	@name	CheckoutPreviewResponse

func NewCheckoutPreviewFromDTO(d *dto.CheckoutPreviewResultDTO) *CheckoutPreviewResponse {
	return &CheckoutPreviewResponse{
		Currency:      d.Currency,
		ItemCount:     d.ItemCount,
		Subtotal:      d.Subtotal,
		DiscountTotal: d.DiscountTotal,
		ShippingTotal: d.ShippingTotal,
		TaxTotal:      d.TaxTotal,
		GrandTotal:    d.GrandTotal,
		Shipping: OrderShippingResponse{
			Postcode:   d.Shipping.Postcode,
			Method:     d.Shipping.Method,
			Provider:   d.Shipping.Provider,
			TariffCode: d.Shipping.TariffCode,
			PointCode:  d.Shipping.PointCode,
			MinDays:    d.Shipping.MinDays,
			MaxDays:    d.Shipping.MaxDays,
		},
	}
}
