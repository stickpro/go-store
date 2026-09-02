package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/models"
)

type ViewedDTO struct {
	Items []ViewedItemDTO
}

type ViewedItemDTO struct {
	ProductID uuid.UUID
	VariantID uuid.UUID
	Name      string
	Slug      string
	Image     *models.ImageDTO
	Price     decimal.Decimal
}
