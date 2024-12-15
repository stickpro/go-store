package category_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"time"
)

type CategoryResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     *string   `json:"description,omitempty"`
	MetaTitle       *string   `json:"meta_title,omitempty"`
	MetaH1          *string   `json:"meta_h1"`
	MetaDescription *string   `json:"meta_description"`
	MetaKeywords    *string   `json:"meta_keywords"`
	IsEnabled       bool      `json:"is_enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func NewFromModel(category *models.Category) CategoryResponse {
	return CategoryResponse{
		ID:              category.ID,
		Name:            category.Name,
		Slug:            category.Slug,
		Description:     pgtypeutils.DecodeText(category.Description),
		MetaTitle:       pgtypeutils.DecodeText(category.MetaTitle),
		MetaH1:          pgtypeutils.DecodeText(category.MetaH1),
		MetaDescription: pgtypeutils.DecodeText(category.MetaDescription),
		MetaKeywords:    pgtypeutils.DecodeText(category.MetaKeyword),
		IsEnabled:       category.IsEnable,
		CreatedAt:       category.CreatedAt.Time,
		UpdatedAt:       category.UpdatedAt.Time,
	}
}
