package mapper

import (
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/repository/repository_collections"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

func MapCollectionToDTO(rows []*repository_collections.GetCollectionWithProductsByIDRow, presets []config.ImagePreset) *dto.WithProductsCollectionDTO { //nolint:dupl
	d := &dto.WithProductsCollectionDTO{
		ID:        rows[0].ID,
		Name:      rows[0].Name,
		Slug:      rows[0].Slug,
		CreatedAt: rows[0].CreatedAt.Time,
	}
	if rows[0].UpdatedAt.Valid {
		d.UpdatedAt = &rows[0].UpdatedAt.Time
	}

	if rows[0].Description.Valid {
		d.Description = pgtypeutils.DecodeText(rows[0].Description)
	}

	for _, row := range rows {
		if !row.ProductID.Valid {
			continue
		}
		d.Products = append(d.Products, &dto.ShortProductDTO{
			ID:             row.VariantID.UUID,
			ProductID:      row.ProductID.UUID,
			Name:           row.ProductName,
			Slug:           row.ProductSlug,
			Model:          row.ProductModel.String,
			PriceRetail:    row.ProductPriceRetail.Decimal,
			PriceBusiness:  row.ProductPriceBusiness.Decimal,
			PriceWholeSale: row.ProductPriceWholesale.Decimal,
			IsEnable:       row.ProductIsEnable.Bool,
			Image:          shortImage(row.ImageID, row.ImagePath, row.ImageWidth, row.ImageHeight, row.ProductName, presets),
		})
	}

	return d
}

func MapCollectionBySlugToDTO(rows []*repository_collections.GetCollectionWithProductsBySlugRow, presets []config.ImagePreset) *dto.WithProductsCollectionDTO { //nolint:dupl
	d := &dto.WithProductsCollectionDTO{
		ID:        rows[0].ID,
		Name:      rows[0].Name,
		Slug:      rows[0].Slug,
		CreatedAt: rows[0].CreatedAt.Time,
	}
	if rows[0].UpdatedAt.Valid {
		d.UpdatedAt = &rows[0].UpdatedAt.Time
	}

	if rows[0].Description.Valid {
		d.Description = pgtypeutils.DecodeText(rows[0].Description)
	}

	for _, row := range rows {
		if !row.ProductID.Valid {
			continue
		}
		d.Products = append(d.Products, &dto.ShortProductDTO{
			ID:             row.VariantID.UUID,
			ProductID:      row.ProductID.UUID,
			Name:           row.ProductName,
			Slug:           row.ProductSlug,
			Model:          row.ProductModel.String,
			PriceRetail:    row.ProductPriceRetail.Decimal,
			PriceBusiness:  row.ProductPriceBusiness.Decimal,
			PriceWholeSale: row.ProductPriceWholesale.Decimal,
			IsEnable:       row.ProductIsEnable.Bool,
			Image:          shortImage(row.ImageID, row.ImagePath, row.ImageWidth, row.ImageHeight, row.ProductName, presets),
		})
	}

	return d
}
