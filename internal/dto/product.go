package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
	"github.com/stickpro/go-store/internal/models"
)

type CreateProductDTO struct {
	Name            string               `json:"name"`
	Model           string               `json:"model"`
	Slug            string               `json:"slug"`
	Description     *string              `json:"description"`
	MetaTitle       *string              `json:"meta_title"`
	MetaH1          *string              `json:"meta_h1"`
	MetaDescription *string              `json:"meta_description"`
	MetaKeyword     *string              `json:"meta_keyword"`
	Sku             *string              `json:"sku"`
	Upc             *string              `json:"upc"`
	Ean             *string              `json:"ean"`
	Jan             *string              `json:"jan"`
	Isbn            *string              `json:"isbn"`
	Mpn             *string              `json:"mpn"`
	Location        *string              `json:"location"`
	Quantity        int64                `json:"quantity"`
	StockStatus     constant.StockStatus `json:"stock_status"`
	Image           *string              `json:"image"`
	ManufacturerID  uuid.NullUUID        `json:"manufacturer_id"`
	Price           decimal.Decimal      `json:"price"`
	Weight          decimal.Decimal      `json:"weight"`
	Length          decimal.Decimal      `json:"length"`
	Width           decimal.Decimal      `json:"width"`
	Height          decimal.Decimal      `json:"height"`
	Subtract        bool                 `json:"subtract"`
	Minimum         int64                `json:"minimum"`
	SortOrder       int32                `json:"sort_order"`
	IsEnable        bool                 `json:"is_enable"`
	MediaIDs        []*uuid.UUID         `json:"media_ids,omitempty"` //nolint:tagliatelle
	CategoryID      uuid.NullUUID        `json:"category_id"`
}

type UpdateProductDTO struct {
	ID              uuid.UUID            `json:"id"`
	Name            string               `json:"name"`
	Model           string               `json:"model"`
	Slug            string               `json:"slug"`
	Description     *string              `json:"description"`
	MetaTitle       *string              `json:"meta_title"`
	MetaH1          *string              `json:"meta_h1"`
	MetaDescription *string              `json:"meta_description"`
	MetaKeyword     *string              `json:"meta_keyword"`
	Sku             *string              `json:"sku"`
	Upc             *string              `json:"upc"`
	Ean             *string              `json:"ean"`
	Jan             *string              `json:"jan"`
	Isbn            *string              `json:"isbn"`
	Mpn             *string              `json:"mpn"`
	Location        *string              `json:"location"`
	Quantity        int64                `json:"quantity"`
	StockStatus     constant.StockStatus `json:"stock_status"`
	Image           *string              `json:"image"`
	ManufacturerID  uuid.NullUUID        `json:"manufacturer_id"`
	Price           decimal.Decimal      `json:"price"`
	Weight          decimal.Decimal      `json:"weight"`
	Length          decimal.Decimal      `json:"length"`
	Width           decimal.Decimal      `json:"width"`
	Height          decimal.Decimal      `json:"height"`
	Subtract        bool                 `json:"subtract"`
	Minimum         int64                `json:"minimum"`
	SortOrder       int32                `json:"sort_order"`
	IsEnable        bool                 `json:"is_enable"`
	MediaIDs        []*uuid.UUID         `json:"media_ids,omitempty"` //nolint:tagliatelle
	CategoryID      uuid.NullUUID        `json:"category_id"`
}

type WithMediumProductDTO struct {
	Product *models.Product  `json:"product"`
	Medium  []*models.Medium `json:"media"` //nolint:tagliatelle
}

type SyncAttributeProductDTO struct {
	ProductID    uuid.UUID   `json:"product_id"`
	AttributeIDs []uuid.UUID `json:"attribute_ids"` //nolint:tagliatelle
}

func RequestToCreateProductDTO(req *product_request.CreateProductRequest) CreateProductDTO {
	var manufacturerID uuid.NullUUID
	var categoryID uuid.NullUUID
	if req.ManufacturerID != nil {
		manufacturerID = uuid.NullUUID{UUID: *req.ManufacturerID, Valid: true}
	} else {
		manufacturerID = uuid.NullUUID{Valid: true}
	}

	if req.CategoryID != nil {
		categoryID = uuid.NullUUID{UUID: *req.CategoryID, Valid: true}
	}

	return CreateProductDTO{
		Name:            req.Name,
		Model:           req.Model,
		Slug:            req.Slug,
		Description:     req.Description,
		MetaTitle:       req.MetaTitle,
		MetaH1:          req.MetaH1,
		MetaDescription: req.MetaDescription,
		MetaKeyword:     req.MetaKeyword,
		Sku:             req.Sku,
		Upc:             req.Upc,
		Ean:             req.Ean,
		Jan:             req.Jan,
		Isbn:            req.Isbn,
		Mpn:             req.Mpn,
		Location:        req.Location,
		Quantity:        req.Quantity,
		StockStatus:     constant.StockStatus(req.StockStatus),
		Image:           req.Image,
		ManufacturerID:  manufacturerID,
		Price:           req.Price,
		Weight:          req.Weight,
		Length:          req.Length,
		Height:          req.Height,
		Width:           req.Width,
		Subtract:        req.Subtract,
		Minimum:         req.Minimum,
		SortOrder:       req.SortOrder,
		IsEnable:        req.IsEnable,
		MediaIDs:        req.MediaIDs,
		CategoryID:      categoryID,
	}
}

func RequestToUpdateProductDTO(req *product_request.UpdateProductRequest, id uuid.UUID) UpdateProductDTO {
	var manufacturerID uuid.NullUUID
	var categoryID uuid.NullUUID
	if req.ManufacturerID != nil {
		manufacturerID = uuid.NullUUID{UUID: *req.ManufacturerID, Valid: true}
	} else {
		manufacturerID = uuid.NullUUID{Valid: true}
	}

	if req.CategoryID != nil {
		categoryID = uuid.NullUUID{UUID: *req.CategoryID, Valid: true}
	}
	return UpdateProductDTO{
		ID:              id,
		Name:            req.Name,
		Model:           req.Model,
		Slug:            req.Slug,
		Description:     req.Description,
		MetaTitle:       req.MetaTitle,
		MetaH1:          req.MetaH1,
		MetaDescription: req.MetaDescription,
		MetaKeyword:     req.MetaKeyword,
		Sku:             req.Sku,
		Upc:             req.Upc,
		Ean:             req.Ean,
		Jan:             req.Jan,
		Isbn:            req.Isbn,
		Mpn:             req.Mpn,
		Location:        req.Location,
		Quantity:        req.Quantity,
		StockStatus:     constant.StockStatus(req.StockStatus),
		Image:           req.Image,
		ManufacturerID:  manufacturerID,
		Price:           req.Price,
		Weight:          req.Weight,
		Length:          req.Length,
		Height:          req.Height,
		Width:           req.Width,
		Subtract:        req.Subtract,
		Minimum:         req.Minimum,
		SortOrder:       req.SortOrder,
		IsEnable:        req.IsEnable,
		MediaIDs:        req.MediaIDs,
		CategoryID:      categoryID,
	}
}
