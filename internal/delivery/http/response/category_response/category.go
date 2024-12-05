package category_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"time"
)

type CategoryResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	MetaTitle       string    `json:"meta_title"`
	MetaH1          string    `json:"meta_h1"`
	MetaDescription string    `json:"meta_description"`
	MetaKeywords    string    `json:"meta_keywords"`
	IsEnabled       bool      `json:"is_enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func NewFromModel(category *models.Category) CategoryResponse {
	return CategoryResponse{
		ID:              category.ID,
		Name:            category.Name,
		Slug:            category.Slug,
		Description:     category.Description.String,
		MetaTitle:       category.MetaTitle.String,
		MetaH1:          category.MetaH1.String,
		MetaDescription: category.MetaDescription.String,
		MetaKeywords:    category.MetaKeyword.String,
		IsEnabled:       category.IsEnable,
		CreatedAt:       category.CreatedAt.Time,
		UpdatedAt:       category.UpdatedAt.Time,
	}
}
