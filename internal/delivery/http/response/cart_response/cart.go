package cart_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/dto"
)

type CartItemResponse struct {
	ProductID   uuid.UUID       `json:"product_id"`
	VariantID   uuid.UUID       `json:"variant_id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	ImageURL    string          `json:"image_url"`
	Price       decimal.Decimal `json:"price"`
	Quantity    int64           `json:"quantity"`
	MaxQuantity int64           `json:"max_quantity"`
	Available   bool            `json:"available"`
} //	@name	CartItemResponse

type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalPrice decimal.Decimal    `json:"total_price"`
} //	@name	CartResponse

func NewFromDTO(d *dto.CartDTO) *CartResponse {
	items := make([]CartItemResponse, len(d.Items))
	for i, item := range d.Items {
		items[i] = CartItemResponse{
			ProductID:   item.ProductID,
			VariantID:   item.VariantID,
			Name:        item.Name,
			Slug:        item.Slug,
			ImageURL:    item.ImageURL,
			Price:       item.Price,
			Quantity:    item.Quantity,
			MaxQuantity: item.MaxQuantity,
			Available:   item.Available,
		}
	}
	return &CartResponse{
		Items:      items,
		TotalPrice: d.TotalPrice,
	}
}
