package models

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ProductID uuid.UUID `json:"product_id"`
	VariantID uuid.UUID `json:"variant_id"`
	Quantity  int       `json:"quantity"`
	AddedAt   time.Time `json:"added_at"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}
