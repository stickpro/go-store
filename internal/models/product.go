package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ShortProduct struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	ProductID uuid.UUID       `db:"product_id" json:"product_id"`
	Name      string          `db:"name" json:"name"`
	Model     string          `db:"model" json:"model"`
	Slug      string          `db:"slug" json:"slug"`
	Image     *ImageDTO       `json:"image,omitempty"`
	Price     decimal.Decimal `db:"price" json:"price"`
	IsEnable  bool            `db:"is_enable" json:"is_enable"`
} //	@name	ShortProduct
