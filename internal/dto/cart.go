package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

type CartOwner struct {
	UserID    *uuid.UUID
	SessionID *uuid.UUID
}
