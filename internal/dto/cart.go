package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/delivery/http/request/cart_request"
	"github.com/stickpro/go-store/internal/models"
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
	Image       *models.ImageDTO
	Price       decimal.Decimal
	Quantity    int64
	MaxQuantity int64
	Available   bool

	// Parcel dimensions of the product, from the products table.
	// Weight is kilograms, dimensions are centimetres; zero means "not set".
	WeightKG decimal.Decimal
	LengthCM decimal.Decimal
	WidthCM  decimal.Decimal
	HeightCM decimal.Decimal
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
