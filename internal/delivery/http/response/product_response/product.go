package product_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/delivery/http/response/medium_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type ProductVariantResponse struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	Slug            string        `json:"slug"`
	CategoryID      uuid.NullUUID `json:"category_id"`
	Description     *string       `json:"description"`
	MetaTitle       *string       `json:"meta_title"`
	MetaH1          *string       `json:"meta_h1"`
	MetaDescription *string       `json:"meta_description"`
	MetaKeyword     *string       `json:"meta_keyword"`
	Image           *string       `json:"image"`
	SortOrder       int32         `json:"sort_order"`
	IsEnable        bool          `json:"is_enable"`
} //	@name	ProductVariantResponse

type ProductResponse struct {
	ID             uuid.UUID              `json:"id"`
	Model          string                 `json:"model"`
	Sku            *string                `json:"sku"`
	Upc            *string                `json:"upc"`
	Ean            *string                `json:"ean"`
	Jan            *string                `json:"jan"`
	Isbn           *string                `json:"isbn"`
	Mpn            *string                `json:"mpn"`
	Location       *string                `json:"location"`
	Quantity       int64                  `json:"quantity"`
	StockStatus    string                 `json:"stock_status"`
	ManufacturerID uuid.NullUUID          `json:"manufacturer_id"`
	Price          decimal.Decimal        `json:"price"`
	Weight         decimal.Decimal        `json:"weight"`
	Length         decimal.Decimal        `json:"length"`
	Width          decimal.Decimal        `json:"width"`
	Height         decimal.Decimal        `json:"height"`
	Subtract       bool                   `json:"subtract"`
	Minimum        int64                  `json:"minimum"`
	SortOrder      int32                  `json:"sort_order"`
	IsEnable       bool                   `json:"is_enable"`
	Variant        ProductVariantResponse `json:"variant"`
} //	@name	ProductResponse

func NewFromModels(product *models.Product, variant *models.ProductVariant) ProductResponse {
	resp := ProductResponse{
		ID:             product.ID,
		Model:          product.Model,
		Sku:            pgtypeutils.DecodeText(product.Sku),
		Upc:            pgtypeutils.DecodeText(product.Upc),
		Ean:            pgtypeutils.DecodeText(product.Ean),
		Jan:            pgtypeutils.DecodeText(product.Jan),
		Isbn:           pgtypeutils.DecodeText(product.Isbn),
		Mpn:            pgtypeutils.DecodeText(product.Mpn),
		Location:       pgtypeutils.DecodeText(product.Location),
		Quantity:       product.Quantity,
		StockStatus:    product.StockStatus.String(),
		ManufacturerID: product.ManufacturerID,
		Price:          product.Price,
		Weight:         product.Weight,
		Length:         product.Length,
		Width:          product.Width,
		Height:         product.Height,
		Subtract:       product.Subtract,
		Minimum:        product.Minimum,
		SortOrder:      product.SortOrder,
		IsEnable:       product.IsEnable,
	}
	if variant != nil {
		resp.Variant = ProductVariantResponse{
			ID:              variant.ID,
			Name:            variant.Name,
			Slug:            variant.Slug,
			CategoryID:      variant.CategoryID,
			Description:     pgtypeutils.DecodeText(variant.Description),
			MetaTitle:       pgtypeutils.DecodeText(variant.MetaTitle),
			MetaH1:          pgtypeutils.DecodeText(variant.MetaH1),
			MetaDescription: pgtypeutils.DecodeText(variant.MetaDescription),
			MetaKeyword:     pgtypeutils.DecodeText(variant.MetaKeyword),
			Image:           pgtypeutils.DecodeText(variant.Image),
			SortOrder:       variant.SortOrder,
			IsEnable:        variant.IsEnable,
		}
	}
	return resp
}

type ProductWithMediumResponse struct {
	Product ProductResponse                  `json:"product"`
	Medium  []medium_response.MediumResponse `json:"medium"`
} //	@name	ProductWithMediumResponse

func NewFromModelsWithMedium(product *models.Product, variant *models.ProductVariant, medium []*models.Medium) ProductWithMediumResponse {
	return ProductWithMediumResponse{
		Product: NewFromModels(product, variant),
		Medium:  medium_response.NewFromModels(medium),
	}
}

func NewVariantsFromModels(variants []*models.ProductVariant) []ProductVariantResponse {
	result := make([]ProductVariantResponse, 0, len(variants))
	for _, v := range variants {
		result = append(result, NewVariantFromModel(v))
	}
	return result
}

func NewVariantFromModel(variant *models.ProductVariant) ProductVariantResponse {
	return ProductVariantResponse{
		ID:              variant.ID,
		Name:            variant.Name,
		Slug:            variant.Slug,
		CategoryID:      variant.CategoryID,
		Description:     pgtypeutils.DecodeText(variant.Description),
		MetaTitle:       pgtypeutils.DecodeText(variant.MetaTitle),
		MetaH1:          pgtypeutils.DecodeText(variant.MetaH1),
		MetaDescription: pgtypeutils.DecodeText(variant.MetaDescription),
		MetaKeyword:     pgtypeutils.DecodeText(variant.MetaKeyword),
		Image:           pgtypeutils.DecodeText(variant.Image),
		SortOrder:       variant.SortOrder,
		IsEnable:        variant.IsEnable,
	}
}

func NewFromProductOnly(product *models.Product) ProductResponse {
	return ProductResponse{
		ID:             product.ID,
		Model:          product.Model,
		Sku:            pgtypeutils.DecodeText(product.Sku),
		Upc:            pgtypeutils.DecodeText(product.Upc),
		Ean:            pgtypeutils.DecodeText(product.Ean),
		Jan:            pgtypeutils.DecodeText(product.Jan),
		Isbn:           pgtypeutils.DecodeText(product.Isbn),
		Mpn:            pgtypeutils.DecodeText(product.Mpn),
		Location:       pgtypeutils.DecodeText(product.Location),
		Quantity:       product.Quantity,
		StockStatus:    product.StockStatus.String(),
		ManufacturerID: product.ManufacturerID,
		Price:          product.Price,
		Weight:         product.Weight,
		Length:         product.Length,
		Width:          product.Width,
		Height:         product.Height,
		Subtract:       product.Subtract,
		Minimum:        product.Minimum,
		SortOrder:      product.SortOrder,
		IsEnable:       product.IsEnable,
	}
}

type ProductAttributeResponse struct {
	Groups []*dto.AttributeGroupWithValuesDTO `json:"groups"`
} //	@name	AttributeGroupsResponse

func NewFromAttributeWithAttributeGroups(attributes []*dto.AttributeGroupWithValuesDTO) ProductAttributeResponse {
	return ProductAttributeResponse{
		Groups: attributes,
	}
}
