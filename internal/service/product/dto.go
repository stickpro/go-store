package product

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_request"
)

type CreateDTO struct {
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
}

type UpdateDTO struct {
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
}

type GetDTO struct {
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
}

func RequestToCreateDTO(req *product_request.CreateProductRequest) CreateDTO {
	var manufacturerID uuid.NullUUID
	if req.ManufacturerID != nil {
		manufacturerID = uuid.NullUUID{UUID: *req.ManufacturerID, Valid: true}
	} else {
		manufacturerID = uuid.NullUUID{Valid: true}
	}
	return CreateDTO{
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
	}
}

func RequestToUpdateDTO(req *product_request.UpdateProductRequest) UpdateDTO {
	var manufacturerID uuid.NullUUID
	if req.ManufacturerID != nil {
		manufacturerID = uuid.NullUUID{UUID: *req.ManufacturerID, Valid: true}
	} else {
		manufacturerID = uuid.NullUUID{Valid: true}
	}
	return UpdateDTO{
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
	}
}
