package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ViewedDTO struct {
	Items []ViewedItemDTO
}

type ViewedItemDTO struct {
	ProductID uuid.UUID
	VariantID uuid.UUID
	Name      string
	Slug      string
	ImageURL  string
	Price     decimal.Decimal
}
