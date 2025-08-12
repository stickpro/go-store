package mapper

import (
	"fmt"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository/repository_collections"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

func MapCollectionToDTO(rows []*repository_collections.GetCollectionWithProductsByIDRow) *dto.WithProductsCollectionDTO {

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
		d.Products = append(d.Products, &models.ShortProduct{
			ID:       row.ProductID.UUID,
			Name:     row.ProductName.String,
			Slug:     row.ProductSlug.String,
			Model:    row.ProductModel.String,
			Price:    row.ProductPrice.Decimal,
			IsEnable: row.ProductIsEnable.Bool,
			Image:    row.ProductImage,
		})
	}

	return d
}

func MapCollectionBySlugToDTO(rows []*repository_collections.GetCollectionWithProductsBySlugRow) *dto.WithProductsCollectionDTO {
	fmt.Println(rows[0].CreatedAt)
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
		d.Products = append(d.Products, &models.ShortProduct{
			ID:       row.ProductID.UUID,
			Name:     row.ProductName.String,
			Slug:     row.ProductSlug.String,
			Model:    row.ProductModel.String,
			Price:    row.ProductPrice.Decimal,
			IsEnable: row.ProductIsEnable.Bool,
			Image:    row.ProductImage,
		})
	}

	return d
}
