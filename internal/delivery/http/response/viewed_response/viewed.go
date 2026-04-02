package viewed_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/dto"
)

type ViewedItemResponse struct {
	ProductID uuid.UUID       `json:"product_id"`
	VariantID uuid.UUID       `json:"variant_id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	ImageURL  string          `json:"image_url"`
	Price     decimal.Decimal `json:"price"`
} //	@name	ViewedItemResponse

type ViewedResponse struct {
	Items []ViewedItemResponse `json:"items"`
} //	@name	ViewedResponse

func NewFromDTO(d *dto.ViewedDTO) *ViewedResponse {
	items := make([]ViewedItemResponse, len(d.Items))
	for i, item := range d.Items {
		items[i] = ViewedItemResponse{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Name:      item.Name,
			Slug:      item.Slug,
			ImageURL:  item.ImageURL,
			Price:     item.Price,
		}
	}
	return &ViewedResponse{Items: items}
}
