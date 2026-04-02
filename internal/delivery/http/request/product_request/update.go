package product_request

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UpdateProductVariantRequest struct {
	ID              uuid.UUID  `json:"id" validate:"required,uuid"`
	Name            string     `json:"name" validate:"omitempty"`
	Slug            string     `json:"slug" validate:"omitempty,slug"`
	CategoryID      *uuid.UUID `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Description     *string    `json:"description,omitempty"`
	MetaTitle       *string    `json:"meta_title,omitempty"`
	MetaH1          *string    `json:"meta_h1,omitempty"`
	MetaDescription *string    `json:"meta_description,omitempty"`
	MetaKeyword     *string    `json:"meta_keyword,omitempty"`
	Image           *string    `json:"image,omitempty"`
	SortOrder       int32      `json:"sort_order"`
	IsEnable        bool       `json:"is_enable"`
} //	@name	UpdateProductVariantRequest

type UpdateProductRequest struct {
	Model          string                      `json:"model" validate:"omitempty"`
	Sku            *string                     `json:"sku,omitempty"`
	Upc            *string                     `json:"upc,omitempty"`
	Ean            *string                     `json:"ean,omitempty"`
	Jan            *string                     `json:"jan,omitempty"`
	Isbn           *string                     `json:"isbn,omitempty"`
	Mpn            *string                     `json:"mpn,omitempty"`
	Location       *string                     `json:"location,omitempty"`
	Quantity       int64                       `json:"quantity"`
	StockStatus    string                      `json:"stock_status"`
	ManufacturerID *uuid.UUID                  `json:"manufacturer_id,omitempty" validate:"omitempty,uuid"`
	PriceRetail    decimal.Decimal             `json:"price_retail" validate:"omitempty,numeric"`
	PriceBusiness  decimal.Decimal             `json:"price_business" validate:"omitempty,numeric"`
	PriceWholesale decimal.Decimal             `json:"price_wholesale" validate:"omitempty,numeric"`
	Weight         decimal.Decimal             `json:"weight" validate:"omitempty,numeric"`
	Length         decimal.Decimal             `json:"length" validate:"omitempty,numeric"`
	Width          decimal.Decimal             `json:"width" validate:"omitempty,numeric"`
	Height         decimal.Decimal             `json:"height" validate:"omitempty,numeric"`
	Subtract       bool                        `json:"subtract"`
	Minimum        int64                       `json:"minimum"`
	SortOrder      int32                       `json:"sort_order"`
	IsEnable       bool                        `json:"is_enable"`
	MediaIDs       []*uuid.UUID                `json:"media_ids,omitempty" validate:"omitempty"` //nolint:tagliatelle
	Variant        UpdateProductVariantRequest `json:"variant" validate:"required"`
} //	@name	UpdateProductRequest
