package manufacturer_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type ManufacturerResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     *string   `json:"description,omitempty"`
	ImagePath       *string   `json:"image_path"`
	MetaTitle       *string   `json:"meta_title,omitempty"`
	MetaH1          *string   `json:"meta_h1"`
	MetaDescription *string   `json:"meta_description"`
	MetaKeywords    *string   `json:"meta_keywords"`
	IsEnabled       bool      `json:"is_enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
} // @name ManufacturerResponse

func NewFromModel(mfc *models.Manufacturer) *ManufacturerResponse {
	var imagePath *string
	if mfc.ImagePath.Valid {
		imagePath = pgtypeutils.DecodeText(mfc.ImagePath)
	}
	return &ManufacturerResponse{
		ID:              mfc.ID,
		Name:            mfc.Name,
		Slug:            mfc.Slug,
		Description:     pgtypeutils.DecodeText(mfc.Description),
		ImagePath:       imagePath,
		MetaTitle:       pgtypeutils.DecodeText(mfc.MetaTitle),
		MetaH1:          pgtypeutils.DecodeText(mfc.MetaH1),
		MetaDescription: pgtypeutils.DecodeText(mfc.MetaDescription),
		MetaKeywords:    pgtypeutils.DecodeText(mfc.MetaKeyword),
		IsEnabled:       mfc.IsEnable,
		CreatedAt:       mfc.CreatedAt.Time,
		UpdatedAt:       mfc.UpdatedAt.Time,
	}
}

func NewPaginatedFromFindRows(
	data *base.FindResponseWithFullPagination[*models.Manufacturer],
) *base.FindResponseWithFullPagination[*ManufacturerResponse] {
	items := make([]*ManufacturerResponse, 0, len(data.Items))
	for _, row := range data.Items {
		items = append(items, &ManufacturerResponse{
			ID:              row.ID,
			Name:            row.Name,
			Slug:            row.Slug,
			Description:     pgtypeutils.DecodeText(row.Description),
			ImagePath:       pgtypeutils.DecodeText(row.ImagePath),
			MetaTitle:       pgtypeutils.DecodeText(row.MetaTitle),
			MetaH1:          pgtypeutils.DecodeText(row.MetaH1),
			MetaDescription: pgtypeutils.DecodeText(row.MetaDescription),
			MetaKeywords:    pgtypeutils.DecodeText(row.MetaKeyword),
			IsEnabled:       row.IsEnable,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		})
	}

	return &base.FindResponseWithFullPagination[*ManufacturerResponse]{
		Items:      items,
		Pagination: data.Pagination,
	}
}
