package product_request

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateProductVariantRequest struct {
	Name            string     `json:"name" validate:"required"`
	Slug            string     `json:"slug" validate:"required,slug"`
	Model           string     `json:"model" validate:"required"`
	CategoryID      *uuid.UUID `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Description     *string    `json:"description,omitempty"`
	MetaTitle       *string    `json:"meta_title,omitempty"`
	MetaH1          *string    `json:"meta_h1,omitempty"`
	MetaDescription *string    `json:"meta_description,omitempty"`
	MetaKeyword     *string    `json:"meta_keyword,omitempty"`
	Image           *string    `json:"image,omitempty"`
	SortOrder       int32      `json:"sort_order"`
	IsEnable        bool       `json:"is_enable"`
} //	@name	CreateProductVariantRequest

type CreateProductRequest struct {
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
	PriceRetail    decimal.Decimal             `json:"price_retail" validate:"numeric"`
	PriceBusiness  decimal.Decimal             `json:"price_business" validate:"numeric"`
	PriceWholeSale decimal.Decimal             `json:"price_wholesale" validate:"numeric"`
	Weight         decimal.Decimal             `json:"weight" validate:"numeric"`
	Length         decimal.Decimal             `json:"length" validate:"numeric"`
	Width          decimal.Decimal             `json:"width" validate:"numeric"`
	Height         decimal.Decimal             `json:"height" validate:"numeric"`
	Subtract       bool                        `json:"subtract"`
	Minimum        int64                       `json:"minimum" validate:"required"`
	SortOrder      int32                       `json:"sort_order"`
	IsEnable       bool                        `json:"is_enable"`
	MediaIDs       []*uuid.UUID                `json:"media_ids,omitempty" validate:"omitempty"` //nolint:tagliatelle
	Variant        CreateProductVariantRequest `json:"variant" validate:"required"`
} //	@name	CreateProductRequest
