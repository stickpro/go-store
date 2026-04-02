package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/delivery/http/request/cart_request"
)

type CartDTO struct {
	Items      []CartItemsDTO
	TotalPrice decimal.Decimal
}

type CartItemsDTO struct {
	ProductID   uuid.UUID
	VariantID   uuid.UUID
	Name        string
	Slug        string
	ImageURL    string
	Price       decimal.Decimal
	Quantity    int64
	MaxQuantity int64
	Available   bool
}

type AddCartItemDTO struct {
	ProductID uuid.UUID
	VariantID uuid.UUID
	Quantity  int64
}

func RequestToAddCartItemDTO(req *cart_request.AddCartItemRequest) AddCartItemDTO {
	return AddCartItemDTO{
		ProductID: req.ProductID,
		VariantID: req.VariantID,
		Quantity:  req.Quantity,
	}
}
