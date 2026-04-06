package product_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/dto"
)

type VariantCategoryResponse struct {
	CategoryID       uuid.UUID `json:"category_id"`
	CategoryName     string    `json:"category_name"`
	CategorySlug     string    `json:"category_slug"`
	CategoryIsEnable bool      `json:"category_is_enable"`
} //	@name	VariantCategoryResponse

func NewVariantCategoriesFromDTO(categories []*dto.VariantCategoryDTO) []VariantCategoryResponse {
	result := make([]VariantCategoryResponse, 0, len(categories))
	for _, c := range categories {
		result = append(result, VariantCategoryResponse{
			CategoryID:       c.CategoryID,
			CategoryName:     c.CategoryName,
			CategorySlug:     c.CategorySlug,
			CategoryIsEnable: c.CategoryIsEnable,
		})
	}
	return result
}
