package category_request

import "github.com/google/uuid"

type UpdateCategoryRequest struct {
	ParentID        *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	Name            string     `json:"name" validate:"omitempty,min=1,max=255"`
	Slug            string     `json:"slug" validate:"omitempty,min=1,max=255"`
	Description     *string    `json:"description" validate:"omitempty,min=1"`
	MetaTitle       *string    `json:"meta_title" validate:"omitempty,min=1"`
	MetaH1          *string    `json:"meta_h1" validate:"omitempty,min=1"`
	MetaDescription *string    `json:"meta_description" validate:"omitempty,min=1"`
	MetaKeyword     *string    `json:"meta_keyword" validate:"omitempty,min=1"`
	IsEnabled       bool       `json:"is_enabled"`
} // @name UpdateCategoryRequest
