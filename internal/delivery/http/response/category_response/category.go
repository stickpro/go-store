package category_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"time"
)

type CategoryResponse struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	Description     *string    `json:"description,omitempty"`
	ImagePath       *string    `json:"image_path"`
	MetaTitle       *string    `json:"meta_title,omitempty"`
	MetaH1          *string    `json:"meta_h1"`
	MetaDescription *string    `json:"meta_description"`
	MetaKeywords    *string    `json:"meta_keywords"`
	ParentID        *uuid.UUID `json:"parent_id"`
	IsEnabled       bool       `json:"is_enabled"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
} // @name CategoryResponse

func NewFromModel(category *models.Category) CategoryResponse {
	var parentID *uuid.UUID
	if category.ParentID.Valid {
		parentID = &category.ParentID.UUID
	}
	var imagePath *string
	if category.ImagePath.Valid {
		str := pgtypeutils.DecodeText(category.ImagePath)
		imagePath = str
	}
	return CategoryResponse{
		ID:              category.ID,
		Name:            category.Name,
		Slug:            category.Slug,
		Description:     pgtypeutils.DecodeText(category.Description),
		ImagePath:       imagePath,
		MetaTitle:       pgtypeutils.DecodeText(category.MetaTitle),
		MetaH1:          pgtypeutils.DecodeText(category.MetaH1),
		MetaDescription: pgtypeutils.DecodeText(category.MetaDescription),
		MetaKeywords:    pgtypeutils.DecodeText(category.MetaKeyword),
		ParentID:        parentID,
		IsEnabled:       category.IsEnable,
		CreatedAt:       category.CreatedAt.Time,
		UpdatedAt:       category.UpdatedAt.Time,
	}
}
