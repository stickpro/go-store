package category_request

import "github.com/google/uuid"

type CreateCategoryRequest struct {
	ParentID        *uuid.UUID `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	Name            string     `json:"name" validate:"required,min=1,max=255"`
	Slug            string     `json:"slug" validate:"required,min=1,max=255,slug"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,min=1"`
	MetaTitle       *string    `json:"meta_title,omitempty" validate:"omitempty,min=1"`
	MetaH1          *string    `json:"meta_h1,omitempty" validate:"omitempty,min=1"`
	MetaDescription *string    `json:"meta_description,omitempty" validate:"omitempty,min=1"`
	MetaKeyword     *string    `json:"meta_keyword,omitempty" validate:"omitempty,min=1"`
	IsEnabled       bool       `json:"is_enabled"`
} // @name CreateCategoryRequest
