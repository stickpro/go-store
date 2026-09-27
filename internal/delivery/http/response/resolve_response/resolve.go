package resolve_response

import (
	"github.com/stickpro/go-store/internal/delivery/http/response/category_response"
	"github.com/stickpro/go-store/internal/delivery/http/response/product_response"
)

const (
	TypeCategory = "category"
	TypeProduct  = "product"
)

type ResolveResponse struct {
	Type string `json:"type"`
	Data any    `json:"data"`
} //	@name	ResolveResponse

func NewFromCategory(category category_response.CategoryResponse) ResolveResponse {
	return ResolveResponse{Type: TypeCategory, Data: category}
}

func NewFromProduct(product product_response.ProductWithMediumResponse) ResolveResponse {
	return ResolveResponse{Type: TypeProduct, Data: product}
}
