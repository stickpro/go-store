package dto

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/models"
)

// ProductUpsertDTO used for create or update product from external system
type ProductUpsertDTO struct {
	ExternalID     string
	Model          string
	Sku            *string
	PriceRetail    decimal.Decimal
	PriceBusiness  decimal.Decimal
	PriceWholesale decimal.Decimal
	Quantity       int64
	StockStatus    constant.StockStatus
	IsEnable       bool
	ManufacturerID uuid.NullUUID
}

type CreateProductVariantDTO struct {
	Name            string
	Slug            string
	Model           string
	CategoryID      uuid.NullUUID
	Description     *string
	MetaTitle       *string
	MetaH1          *string
	MetaDescription *string
	MetaKeyword     *string
	SortOrder       int32
	IsEnable        bool
}

type CreateProductDTO struct {
	Sku            *string
	Upc            *string
	Ean            *string
	Jan            *string
	Isbn           *string
	Mpn            *string
	Location       *string
	Quantity       int64
	Image          *string
	StockStatus    constant.StockStatus
	ManufacturerID uuid.NullUUID
	PriceRetail    decimal.Decimal
	PriceBusiness  decimal.Decimal
	PriceWholesale decimal.Decimal
	Weight         decimal.Decimal
	Length         decimal.Decimal
	Width          decimal.Decimal
	Height         decimal.Decimal
	Subtract       bool
	Minimum        int64
	SortOrder      int32
	IsEnable       bool
	MediaIDs       []*uuid.UUID
}

type UpdateProductVariantDTO struct {
	ID              uuid.UUID
	Name            string
	Slug            string
	Model           string
	CategoryID      uuid.NullUUID
	Description     *string
	MetaTitle       *string
	MetaH1          *string
	MetaDescription *string
	MetaKeyword     *string
	Image           *string
	SortOrder       int32
	IsEnable        bool
}

type UpdateProductDTO struct {
	ID             uuid.UUID
	Sku            *string
	Upc            *string
	Ean            *string
	Jan            *string
	Isbn           *string
	Mpn            *string
	Location       *string
	Quantity       int64
	Image          *string
	StockStatus    constant.StockStatus
	ManufacturerID uuid.NullUUID
	PriceRetail    decimal.Decimal
	PriceBusiness  decimal.Decimal
	PriceWholeSale decimal.Decimal
	Weight         decimal.Decimal
	Length         decimal.Decimal
	Width          decimal.Decimal
	Height         decimal.Decimal
	Subtract       bool
	Minimum        int64
	SortOrder      int32
	IsEnable       bool
	MediaIDs       []*uuid.UUID
}

type ProductWithMediaDTO struct {
	Product *models.Product        `json:"product"`
	Variant *models.ProductVariant `json:"variant"`
	Medium  []*models.Medium       `json:"media"` //nolint:tagliatelle
}

// EnrichedVariantDTO combines variant data with product-level fields for indexing and enriched responses
type EnrichedVariantDTO struct {
	*models.ProductVariant
	PriceRetail    decimal.Decimal      `json:"price_retail"`
	PriceBusiness  decimal.Decimal      `json:"price_business"`
	PriceWholesale decimal.Decimal      `json:"price_wholesale"`
	ManufacturerID uuid.NullUUID        `json:"manufacturer_id"`
	StockStatus    constant.StockStatus `json:"stock_status"`
}

type SyncVariantCategoriesDTO struct {
	VariantID   uuid.UUID   `json:"variant_id"`
	CategoryIDs []uuid.UUID `json:"category_ids"` //nolint:tagliatelle
}

type VariantCategoryDTO struct {
	CategoryID       uuid.UUID `json:"category_id"`
	CategoryName     string    `json:"category_name"`
	CategorySlug     string    `json:"category_slug"`
	CategoryIsEnable bool      `json:"category_is_enable"`
}

type SyncAttributeProductDTO struct {
	ProductID         uuid.UUID   `json:"product_id"`
	AttributeValueIDs []uuid.UUID `json:"attribute_value_ids"` //nolint:tagliatelle
}

func RequestToCreateProductDTO(req *product_request.CreateProductRequest) CreateProductDTO {
	var manufacturerID uuid.NullUUID
	if req.ManufacturerID != nil {
		manufacturerID = uuid.NullUUID{UUID: *req.ManufacturerID, Valid: true}
	}

	return CreateProductDTO{
		Sku:            req.Sku,
		Upc:            req.Upc,
		Ean:            req.Ean,
		Jan:            req.Jan,
		Isbn:           req.Isbn,
		Mpn:            req.Mpn,
		Location:       req.Location,
		Quantity:       req.Quantity,
		StockStatus:    constant.StockStatus(req.StockStatus),
		ManufacturerID: manufacturerID,
		PriceRetail:    req.PriceRetail,
		PriceBusiness:  req.PriceBusiness,
		PriceWholesale: req.PriceWholesale,
		Weight:         req.Weight,
		Length:         req.Length,
		Height:         req.Height,
		Width:          req.Width,
		Subtract:       req.Subtract,
		Minimum:        req.Minimum,
		SortOrder:      req.SortOrder,
		IsEnable:       req.IsEnable,
		MediaIDs:       req.MediaIDs,
		Image:          req.Image,
	}
}

func RequestToUpdateProductDTO(req *product_request.UpdateProductRequest, id uuid.UUID) UpdateProductDTO {
	var manufacturerID uuid.NullUUID
	if req.ManufacturerID != nil {
		manufacturerID = uuid.NullUUID{UUID: *req.ManufacturerID, Valid: true}
	}

	return UpdateProductDTO{
		ID:             id,
		Sku:            req.Sku,
		Upc:            req.Upc,
		Ean:            req.Ean,
		Jan:            req.Jan,
		Isbn:           req.Isbn,
		Mpn:            req.Mpn,
		Location:       req.Location,
		Quantity:       req.Quantity,
		StockStatus:    constant.StockStatus(req.StockStatus),
		ManufacturerID: manufacturerID,
		PriceRetail:    req.PriceRetail,
		PriceBusiness:  req.PriceBusiness,
		PriceWholeSale: req.PriceWholesale,
		Weight:         req.Weight,
		Length:         req.Length,
		Height:         req.Height,
		Width:          req.Width,
		Subtract:       req.Subtract,
		Minimum:        req.Minimum,
		SortOrder:      req.SortOrder,
		IsEnable:       req.IsEnable,
		MediaIDs:       req.MediaIDs,
	}
}

type ShortProductDTO struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	Name           string
	Model          string
	Slug           string
	Image          pgtype.Text
	PriceRetail    decimal.Decimal
	PriceBusiness  decimal.Decimal
	PriceWholeSale decimal.Decimal
	IsEnable       bool
}
