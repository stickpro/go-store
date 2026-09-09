package order_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/base"
)

type OrderShippingResponse struct {
	CityID     *uuid.UUID `json:"city_id"`
	CityName   string     `json:"city_name"`
	Address    string     `json:"address"`
	Postcode   *string    `json:"postcode"`
	Recipient  string     `json:"recipient"`
	Method     *string    `json:"method"`
	Provider   *string    `json:"provider"`
	TariffCode *string    `json:"tariff_code"`
	PointCode  *string    `json:"point_code"`
	MinDays    *int32     `json:"min_days"`
	MaxDays    *int32     `json:"max_days"`
} //	@name	OrderShippingResponse

type OrderItemResponse struct {
	ProductID *uuid.UUID      `json:"product_id"`
	VariantID *uuid.UUID      `json:"variant_id"`
	Sku       *string         `json:"sku"`
	Name      string          `json:"name"`
	Slug      *string         `json:"slug"`
	ImagePath *string         `json:"image_path"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	Quantity  int64           `json:"quantity"`
	LineTotal decimal.Decimal `json:"line_total"`
} //	@name	OrderItemResponse

type OrderResponse struct {
	ID            uuid.UUID             `json:"id"`
	Number        int64                 `json:"number"`
	Status        string                `json:"status"`
	PaymentStatus string                `json:"payment_status"`
	PaymentMethod *string               `json:"payment_method"`
	Currency      string                `json:"currency"`
	Email         string                `json:"email"`
	Phone         *string               `json:"phone"`
	Shipping      OrderShippingResponse `json:"shipping"`
	Items         []OrderItemResponse   `json:"items"`
	Subtotal      decimal.Decimal       `json:"subtotal"`
	DiscountTotal decimal.Decimal       `json:"discount_total"`
	ShippingTotal decimal.Decimal       `json:"shipping_total"`
	TaxTotal      decimal.Decimal       `json:"tax_total"`
	GrandTotal    decimal.Decimal       `json:"grand_total"`
	Comment       *string               `json:"comment"`
	CreatedAt     time.Time             `json:"created_at"`
	PaidAt        *time.Time            `json:"paid_at"`
	CancelledAt   *time.Time            `json:"cancelled_at"`
} //	@name	OrderResponse

func NewFromDTO(d *dto.OrderDTO) *OrderResponse {
	items := make([]OrderItemResponse, 0, len(d.Items))
	for _, it := range d.Items {
		items = append(items, OrderItemResponse{
			ProductID: it.ProductID,
			VariantID: it.VariantID,
			Sku:       it.Sku,
			Name:      it.Name,
			Slug:      it.Slug,
			ImagePath: it.ImagePath,
			UnitPrice: it.UnitPrice,
			Quantity:  it.Quantity,
			LineTotal: it.LineTotal,
		})
	}

	return &OrderResponse{
		ID:            d.ID,
		Number:        d.Number,
		Status:        d.Status,
		PaymentStatus: d.PaymentStatus,
		PaymentMethod: d.PaymentMethod,
		Currency:      d.Currency,
		Email:         d.Email,
		Phone:         d.Phone,
		Shipping: OrderShippingResponse{
			CityID:     d.Shipping.CityID,
			CityName:   d.Shipping.CityName,
			Address:    d.Shipping.Address,
			Postcode:   d.Shipping.Postcode,
			Recipient:  d.Shipping.Recipient,
			Method:     d.Shipping.Method,
			Provider:   d.Shipping.Provider,
			TariffCode: d.Shipping.TariffCode,
			PointCode:  d.Shipping.PointCode,
			MinDays:    d.Shipping.MinDays,
			MaxDays:    d.Shipping.MaxDays,
		},
		Items:         items,
		Subtotal:      d.Subtotal,
		DiscountTotal: d.DiscountTotal,
		ShippingTotal: d.ShippingTotal,
		TaxTotal:      d.TaxTotal,
		GrandTotal:    d.GrandTotal,
		Comment:       d.Comment,
		CreatedAt:     d.CreatedAt,
		PaidAt:        d.PaidAt,
		CancelledAt:   d.CancelledAt,
	}
}

// NewPaginated maps a page of order DTOs to the response contract.
func NewPaginated(
	data *base.FindResponseWithFullPagination[*dto.OrderDTO],
) *base.FindResponseWithFullPagination[*OrderResponse] {
	items := make([]*OrderResponse, 0, len(data.Items))
	for _, o := range data.Items {
		items = append(items, NewFromDTO(o))
	}
	return &base.FindResponseWithFullPagination[*OrderResponse]{
		Items:      items,
		Pagination: data.Pagination,
	}
}
