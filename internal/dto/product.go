package dto

import (
	"github.com/google/uuid"
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
	Price          decimal.Decimal
	Quantity       int64
	StockStatus    constant.StockStatus
	IsEnable       bool
	ManufacturerID uuid.NullUUID
}

type CreateProductVariantDTO struct {
	Name            string
	Slug            string
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

type CreateProductDTO struct {
	Model          string
	Sku            *string
	Upc            *string
	Ean            *string
	Jan            *string
	Isbn           *string
	Mpn            *string
	Location       *string
	Quantity       int64
	StockStatus    constant.StockStatus
	ManufacturerID uuid.NullUUID
	Price          decimal.Decimal
	Weight         decimal.Decimal
	Length         decimal.Decimal
	Width          decimal.Decimal
	Height         decimal.Decimal
	Subtract       bool
	Minimum        int64
	SortOrder      int32
	IsEnable       bool
	MediaIDs       []*uuid.UUID
	Variant        CreateProductVariantDTO
}

type UpdateProductVariantDTO struct {
	ID              uuid.UUID
	Name            string
	Slug            string
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
	Model          string
	Sku            *string
	Upc            *string
	Ean            *string
	Jan            *string
	Isbn           *string
	Mpn            *string
	Location       *string
	Quantity       int64
	StockStatus    constant.StockStatus
	ManufacturerID uuid.NullUUID
	Price          decimal.Decimal
	Weight         decimal.Decimal
	Length         decimal.Decimal
	Width          decimal.Decimal
	Height         decimal.Decimal
	Subtract       bool
	Minimum        int64
	SortOrder      int32
	IsEnable       bool
	MediaIDs       []*uuid.UUID
	Variant        UpdateProductVariantDTO
}

type ProductWithMediaDTO struct {
	Product *models.Product        `json:"product"`
	Variant *models.ProductVariant `json:"variant"`
	Medium  []*models.Medium       `json:"media"` //nolint:tagliatelle
}

// EnrichedVariantDTO combines variant data with product-level fields for indexing and enriched responses
type EnrichedVariantDTO struct {
	*models.ProductVariant
	Price          decimal.Decimal      `json:"price"`
	ManufacturerID uuid.NullUUID        `json:"manufacturer_id"`
	StockStatus    constant.StockStatus `json:"stock_status"`
	Model          string               `json:"model"`
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

	var categoryID uuid.NullUUID
	if req.Variant.CategoryID != nil {
		categoryID = uuid.NullUUID{UUID: *req.Variant.CategoryID, Valid: true}
	}

	return CreateProductDTO{
		Model:          req.Model,
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
		Price:          req.Price,
		Weight:         req.Weight,
		Length:         req.Length,
		Height:         req.Height,
		Width:          req.Width,
		Subtract:       req.Subtract,
		Minimum:        req.Minimum,
		SortOrder:      req.SortOrder,
		IsEnable:       req.IsEnable,
		MediaIDs:       req.MediaIDs,
		Variant: CreateProductVariantDTO{
			Name:            req.Variant.Name,
			Slug:            req.Variant.Slug,
			CategoryID:      categoryID,
			Description:     req.Variant.Description,
			MetaTitle:       req.Variant.MetaTitle,
			MetaH1:          req.Variant.MetaH1,
			MetaDescription: req.Variant.MetaDescription,
			MetaKeyword:     req.Variant.MetaKeyword,
			Image:           req.Variant.Image,
			SortOrder:       req.Variant.SortOrder,
			IsEnable:        req.Variant.IsEnable,
		},
	}
}

func RequestToUpdateProductDTO(req *product_request.UpdateProductRequest, id uuid.UUID) UpdateProductDTO {
	var manufacturerID uuid.NullUUID
	if req.ManufacturerID != nil {
		manufacturerID = uuid.NullUUID{UUID: *req.ManufacturerID, Valid: true}
	}

	var categoryID uuid.NullUUID
	if req.Variant.CategoryID != nil {
		categoryID = uuid.NullUUID{UUID: *req.Variant.CategoryID, Valid: true}
	}

	return UpdateProductDTO{
		ID:             id,
		Model:          req.Model,
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
		Price:          req.Price,
		Weight:         req.Weight,
		Length:         req.Length,
		Height:         req.Height,
		Width:          req.Width,
		Subtract:       req.Subtract,
		Minimum:        req.Minimum,
		SortOrder:      req.SortOrder,
		IsEnable:       req.IsEnable,
		MediaIDs:       req.MediaIDs,
		Variant: UpdateProductVariantDTO{
			ID:              req.Variant.ID,
			Name:            req.Variant.Name,
			Slug:            req.Variant.Slug,
			CategoryID:      categoryID,
			Description:     req.Variant.Description,
			MetaTitle:       req.Variant.MetaTitle,
			MetaH1:          req.Variant.MetaH1,
			MetaDescription: req.Variant.MetaDescription,
			MetaKeyword:     req.Variant.MetaKeyword,
			Image:           req.Variant.Image,
			SortOrder:       req.Variant.SortOrder,
			IsEnable:        req.Variant.IsEnable,
		},
	}
}
