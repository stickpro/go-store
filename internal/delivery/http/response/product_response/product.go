package product_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type ProductResponse struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Model           string          `json:"model"`
	Slug            string          `json:"slug"`
	Description     *string         `json:"description"`
	MetaTitle       *string         `json:"meta_title"`
	MetaH1          *string         `json:"meta_h1"`
	MetaDescription *string         `json:"meta_description"`
	MetaKeyword     *string         `json:"meta_keyword"`
	Sku             *string         `json:"sku"`
	Upc             *string         `json:"upc"`
	Ean             *string         `json:"ean"`
	Jan             *string         `json:"jan"`
	Isbn            *string         `json:"isbn"`
	Mpn             *string         `json:"mpn"`
	Location        string          `json:"location"`
	Quantity        int64           `json:"quantity"`
	StockStatus     string          `json:"stock_status"`
	Image           *string         `json:"image"`
	ManufacturerID  uuid.NullUUID   `json:"manufacturer_id"`
	Price           decimal.Decimal `json:"price"`
	Weight          decimal.Decimal `json:"weight"`
	Length          decimal.Decimal `json:"length"`
	Width           decimal.Decimal `json:"width"`
	Height          decimal.Decimal `json:"height"`
	Subtract        bool            `json:"subtract"`
	Minimum         int64           `json:"minimum"`
	SortOrder       int32           `json:"sort_order"`
	IsEnable        bool            `json:"is_enable"`
}

func NewFromModel(product *models.Product) ProductResponse {
	return ProductResponse{
		ID:              product.ID,
		Name:            product.Name,
		Slug:            product.Slug,
		Description:     pgtypeutils.DecodeText(product.Description),
		MetaTitle:       pgtypeutils.DecodeText(product.MetaTitle),
		MetaH1:          pgtypeutils.DecodeText(product.MetaH1),
		MetaDescription: pgtypeutils.DecodeText(product.MetaDescription),
		MetaKeyword:     pgtypeutils.DecodeText(product.MetaKeyword),
		Sku:             pgtypeutils.DecodeText(product.Sku),
		Upc:             pgtypeutils.DecodeText(product.Upc),
		Ean:             pgtypeutils.DecodeText(product.Ean),
		Jan:             pgtypeutils.DecodeText(product.Jan),
		Isbn:            pgtypeutils.DecodeText(product.Isbn),
		Mpn:             pgtypeutils.DecodeText(product.Mpn),
		Location:        product.Location.String,
		Quantity:        product.Quantity,
		StockStatus:     product.StockStatus.String(),
		Image:           pgtypeutils.DecodeText(product.Image),
		ManufacturerID:  product.ManufacturerID,
		Price:           product.Price,
		Weight:          product.Weight,
		Length:          product.Length,
		Width:           product.Width,
		Height:          product.Height,
		Subtract:        product.Subtract,
		Minimum:         product.Minimum,
		SortOrder:       product.SortOrder,
		IsEnable:        product.IsEnable,
	}
}
