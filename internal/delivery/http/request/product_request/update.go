package product_request

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UpdateProductRequest struct {
	Name            string          `json:"name" validate:"omitempty"`
	Model           string          `json:"model" validate:"omitempty"`
	Slug            string          `json:"slug" validate:"omitempty,slug"`
	Description     *string         `json:"description,omitempty" validate:"omitempty"`
	MetaTitle       *string         `json:"meta_title,omitempty" validate:"omitempty"`
	MetaH1          *string         `json:"meta_h1,omitempty" validate:"omitempty"`
	MetaDescription *string         `json:"meta_description,omitempty" validate:"omitempty"`
	MetaKeyword     *string         `json:"meta_keyword,omitempty" validate:"omitempty"`
	Sku             *string         `json:"sku,omitempty" validate:"omitempty"`
	Upc             *string         `json:"upc,omitempty" validate:"omitempty"`
	Ean             *string         `json:"ean,omitempty" validate:"omitempty"`
	Jan             *string         `json:"jan,omitempty" validate:"omitempty"`
	Isbn            *string         `json:"isbn,omitempty" validate:"omitempty"`
	Mpn             *string         `json:"mpn,omitempty" validate:"omitempty"`
	Location        *string         `json:"location,omitempty" validate:"omitempty"`
	Quantity        int64           `json:"quantity" validate:"omitempty"`
	StockStatus     string          `json:"stock_status" validate:"omitempty"`
	Image           *string         `json:"image" validate:"omitempty"`
	ManufacturerID  *uuid.UUID      `json:"manufacturer_id,omitempty" validate:"omitempty,uuid"`
	Price           decimal.Decimal `json:"price" validate:"omitempty,numeric"`
	Weight          decimal.Decimal `json:"weight" validate:"omitempty,numeric"`
	Length          decimal.Decimal `json:"length" validate:"omitempty,numeric"`
	Width           decimal.Decimal `json:"width" validate:"omitempty,numeric"`
	Height          decimal.Decimal `json:"height" validate:"omitempty,numeric"`
	Subtract        bool            `json:"subtract" validate:"omitempty,boolean"`
	Minimum         int64           `json:"minimum" validate:"omitempty"`
	SortOrder       int32           `json:"sort_order" validate:"omitempty"`
	IsEnable        bool            `json:"is_enable" validate:"omitempty,boolean"`
	MediaIDs        []*uuid.UUID    `json:"media_ids,omitempty" validate:"omitempty"` //nolint:tagliatelle
} // @name UpdateProductRequest
